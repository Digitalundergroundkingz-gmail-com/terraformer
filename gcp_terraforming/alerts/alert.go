package alerts

import (
	"context"
	"log"
	"os"

	"waze/terraform/gcp_terraforming/gcp_generator"
	"waze/terraform/terraform_utils"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/monitoring/apiv3"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

var alertsIgnoreKey = map[string]bool{
	"^creation_record":       true,
	"^name$":                  true,
	"^id$":                    true,
	"^conditions.[0-9].name$": true,
}

var alertsAllowEmptyValues = map[string]bool{}

var alertsAdditionalFields = map[string]string{}

type AlertsGenerator struct {
	gcp_generator.BasicGenerator
}

func (AlertsGenerator) createResources(alertIterator *monitoring.AlertPolicyIterator) []terraform_utils.TerraformResource {
	resources := []terraform_utils.TerraformResource{}
	for {
		alert, err := alertIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println("error with alert:", err)
			continue
		}
		resources = append(resources, terraform_utils.NewTerraformResource(
			alert.Name,
			alert.Name,
			"google_monitoring_alert_policy",
			"google",
			nil,
			map[string]string{
				"name": alert.Name,
			},
		))
	}
	return resources
}

func (g AlertsGenerator) Generate(zone string) ([]terraform_utils.TerraformResource, map[string]terraform_utils.ResourceMetaData, error) {
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	ctx := context.Background()
	req := &monitoringpb.ListAlertPoliciesRequest{
		Name: "projects/" + project,
	}

	client, err := monitoring.NewAlertPolicyClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	alertIterator := client.ListAlertPolicies(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	resources := g.createResources(alertIterator)
	metadata := terraform_utils.NewResourcesMetaData(resources, alertsIgnoreKey, alertsAllowEmptyValues, alertsAdditionalFields)
	return resources, metadata, nil

}