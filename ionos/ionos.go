// Package ionos implements a cert-manager ACME DNS01 webhook solver for IONOS DNS.
package ionos

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	corev1 "k8s.io/api/core/v1"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"sync"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	acme "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/miekg/dns"
	"k8s.io/client-go/rest"
)

type ionosDNSProviderConfig struct {
	PublicKeySecretRef corev1.SecretKeySelector `json:"publicKeySecretRef"`
	SecretKeySecretRef corev1.SecretKeySelector `json:"secretKeySecretRef"`
	ZoneName           string                   `json:"zoneName"`
	ApiUrl             string                   `json:"apiUrl"`
}

type Config struct {
	ApiKey   string
	ZoneName string
	ApiUrl   string
}

type ionosSolver struct {
	context     context.Context
	ionosClient Client
	client      *kubernetes.Clientset
	name        string
	server      *dns.Server
	sync.RWMutex
}

func (e *ionosSolver) Name() string {
	return e.name
}

func (e *ionosSolver) Present(ch *acme.ChallengeRequest) error {
	klog.V(6).Infof("call function Present: namespace=%s, zone=%s, fqdn=%s", ch.ResourceNamespace, ch.ResolvedZone, ch.ResolvedFQDN)

	config, err := e.clientInit(ch)

	if err != nil {
		return fmt.Errorf("unable to init client `%s`; %v", ch.ResourceNamespace, err)
	}

	return e.addTxtRecord(config.ZoneName, ch)
}

func (e *ionosSolver) CleanUp(ch *acme.ChallengeRequest) error {

	config, err := e.clientInit(ch)

	if err != nil {
		return fmt.Errorf("unable to init client `%s`; %v", ch.ResourceNamespace, err)
	}

	err = e.removeTxtRecord(config.ZoneName, ch)

	if err != nil {
		return fmt.Errorf("cleanup not possible %v", err)
	}

	return nil
}

func (e *ionosSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	go func() {
		if e.server == nil {
			return
		}
		if err := e.server.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
	}()

	k8sClient, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	e.client = k8sClient

	return nil
}
func (e *ionosSolver) clientInit(ch *acme.ChallengeRequest) (Config, error) {
	var config Config

	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return config, err
	}
	config.ZoneName = cfg.ZoneName
	config.ApiUrl = cfg.ApiUrl

	if config.ApiUrl == "" {
		config.ApiUrl = "https://api.hosting.ionos.com/dns/v1"
	}

	secretKey, err := e.getSecret(cfg.SecretKeySecretRef, ch.ResourceNamespace)
	if err != nil {
		return config, err
	}
	publicKey, err := e.getSecret(cfg.PublicKeySecretRef, ch.ResourceNamespace)
	if err != nil {
		return config, err
	}
	config.ApiKey = publicKey + "." + secretKey

	e.ionosClient.SetConfig(e.context, &config)

	// Get ZoneName by api search if not provided by config
	if config.ZoneName == "" {
		foundZone, err := e.searchZoneName(ch.ResolvedZone)
		if err != nil {
			return config, err
		}
		config.ZoneName = foundZone
	}

	return config, nil
}

func (e *ionosSolver) searchZoneName(searchZone string) (string, error) {
	parts := strings.Split(searchZone, ".")
	parts = parts[:len(parts)-1]
	var lastErr error
	for i := 0; i <= len(parts)-2; i++ {
		name := strings.Join(parts[i:], ".")
		zoneId, err := e.ionosClient.GetZoneIdByName(e.context, name)
		if err != nil {
			klog.Warningf("failed to query zone %q: %v", name, err)
			lastErr = err
			continue
		}
		if zoneId != "" {
			klog.Infof("Found ID with ZoneName: %s", name)
			return name, nil
		}
	}
	if lastErr != nil {
		return "", fmt.Errorf("unable to find ionos dns zone with %q, last API error: %v", searchZone, lastErr)
	}
	return "", fmt.Errorf("unable to find ionos dns zone with: %s", searchZone)
}

func (e *ionosSolver) getSecret(selector corev1.SecretKeySelector, namespace string) (string, error) {
	secret, err := e.client.CoreV1().Secrets(namespace).Get(e.context, selector.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to load secret %q; %v", namespace+"/"+selector.Name, err)
	}

	if data, ok := secret.Data[selector.Key]; ok {
		return string(data), nil
	}

	return "", fmt.Errorf("key not found %q in secret '%s/%s'", selector.Key, namespace, selector.Name)
}

func (e *ionosSolver) removeTxtRecord(zoneName string, ch *acme.ChallengeRequest) error {

	zoneId, err := e.ionosClient.GetZoneIdByName(e.context, zoneName)

	if err != nil {
		return fmt.Errorf("unable to find id for zone name `%s`; %v", zoneName, err)
	}

	name := util.UnFqdn(ch.ResolvedFQDN)

	recordId, err := e.ionosClient.GetRecordIdByName(e.context, zoneId, name)

	if err != nil {
		// Record may already be deleted — treat as success for idempotent cleanup
		klog.Warningf("record %q not found, may already be deleted: %v", name, err)
		return nil
	}

	err = e.ionosClient.DeleteRecord(e.context, zoneId, recordId)

	if err != nil {
		return fmt.Errorf("unable to delete record with id `%s`; %v", recordId, err)
	}

	return nil
}

func (e *ionosSolver) addTxtRecord(zoneName string, ch *acme.ChallengeRequest) error {

	name := util.UnFqdn(ch.ResolvedFQDN)
	content := ch.Key
	zoneId, err := e.ionosClient.GetZoneIdByName(e.context, zoneName)

	if err != nil {
		return fmt.Errorf("unable to find id for zone name `%s`; %v", zoneName, err)
	}

	request := RecordCreateRequest{}
	request = append(request, RecordCreate{
		Name:     name,
		Type:     "TXT",
		Content:  content,
		Ttl:      120,
		Disabled: false,
	})

	err = e.ionosClient.AddRecord(e.context, zoneId, request)

	if err != nil {
		return fmt.Errorf("unable to add TXT record; %v", err)
	}
	klog.Infof("Added TXT record successful")
	return nil
}

func loadConfig(cfgJSON *extapi.JSON) (ionosDNSProviderConfig, error) {
	cfg := ionosDNSProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

func New() webhook.Solver {
	return &ionosSolver{
		context:     context.Background(),
		name:        "ionos",
		ionosClient: NewClient(),
	}
}

func NewMock(port string) webhook.Solver {
	e := &ionosSolver{
		context:     context.Background(),
		name:        "ionos",
		ionosClient: NewMockClient(),
	}

	e.server = &dns.Server{
		Addr:    ":" + port,
		Net:     "udp",
		Handler: dns.HandlerFunc(e.handleDNSRequest),
	}
	return e
}
