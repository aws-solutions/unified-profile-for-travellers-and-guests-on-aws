package servicediscovery

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	sd "github.com/aws/aws-sdk-go/service/servicediscovery"
)

type Config struct {
	NamespaceName string
	Client        *sd.ServiceDiscovery
}

func Init(namespaceName string) Config {
	mySession := session.Must(session.NewSession())
	return Config{
		NamespaceName: namespaceName,
		Client:        sd.New(mySession),
	}
}

func (cfg Config) Url(serviceName string) (string, error) {
	input := servicediscovery.DiscoverInstancesInput{
		NamespaceName: aws.String(cfg.NamespaceName),
		ServiceName:   aws.String(serviceName),
	}
	res, err := cfg.Client.DiscoverInstances(&input)
	if err != nil {
		log.Printf("[CORE][SERVICE DISCOVERY] Error during service discovery for namespace %v", err)
		return "", err
	}
	log.Printf("[CORE][SERVICE DISCOVERY] Found Services for namespace %v", cfg.NamespaceName)
	if len(res.Instances) == 0 {
		log.Printf("[CORE][SERVICE DISCOVERY] No service found in namespace %v", cfg.NamespaceName)
		return "", errors.New("No service found")
	}
	return *res.Instances[0].Attributes["url"], nil
}
