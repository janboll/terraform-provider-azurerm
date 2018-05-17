package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/monitor/mgmt/2017-05-01-preview/insights"
	"github.com/hashicorp/terraform/helper/schema"
)

func diagnosticLoggingSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},

				"storage_account_id": {
					Type:     schema.TypeString,
					Optional: true,
				},

				"event_hub_name": {
					Type:     schema.TypeString,
					Optional: true,
				},

				"event_hub_authorization_rule_id": {
					Type:     schema.TypeString,
					Optional: true,
				},

				"workspace_id": {
					Type:     schema.TypeString,
					Optional: true,
				},

				"metric_settings": {
					Type:     schema.TypeSet,
					Optional: true,
					MinItems: 1,
					MaxItems: 16,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"category": {
								Type:     schema.TypeString,
								Required: true,
							},
							"enabled": {
								Type:     schema.TypeBool,
								Required: true,
							},
							"retention_days": {
								Type:     schema.TypeInt,
								Required: true,
							},
						},
					},
				},

				"logs_settings": {
					Type:     schema.TypeSet,
					Optional: true,
					MinItems: 1,
					MaxItems: 16,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"category": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"enabled": {
								Type:     schema.TypeBool,
								Computed: true,
							},
							"retention_days": {
								Type:     schema.TypeInt,
								Computed: true,
							},
						},
					},
				},
			},
		},
	}
}

func createOrDeleteDiagnosticLogging(resourceId string, diagSettings interface{}, meta interface{}) error {
	currentDiagSettings := listExistingConfiguration(resourceId, meta)
	settingsList := diagSettings.(*schema.Set).List()

	settings := settingsList[0].(map[string]interface{})

	if len(*currentDiagSettings) > 0 && len(settingsList) == 0 {
		return deleteDiagnosticLogging(resourceId, currentDiagSettings, meta)
	}
	return createDiagNosticLogging(resourceId, settings, meta)
}

func createDiagNosticLogging(resourceId string, settings map[string]interface{}, meta interface{}) error {
	client := meta.(*ArmClient).monitorDiagnosticSettingsClient
	ctx := meta.(*ArmClient).StopContext
	log.Printf("[INFO] preparing arguments for Azure ARM KeyVault creation.")

	name := settings["name"].(string)
	diagnosticSettings := expandDiagnosticSettings(settings)

	_, err := client.CreateOrUpdate(
		ctx,
		resourceId,
		insights.DiagnosticSettingsResource{
			Name:               &name,
			DiagnosticSettings: &diagnosticSettings,
		},
		name)
	if err != nil {
		return err
	}

	read, err := client.Get(ctx, resourceId, name)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read Diagnostic Settings")
	}

	return nil
}

func listExistingConfiguration(resourceId string, meta interface{}) *[]string {
	client := meta.(*ArmClient).monitorDiagnosticSettingsClient
	ctx := meta.(*ArmClient).StopContext
	returnSlice := []string{}

	read, err := client.List(ctx, resourceId)

	if err != nil {
		fmt.Println("lol")
	}

	for _, element := range *read.Value {
		returnSlice = append(returnSlice, *element.Name)
	}

	return &returnSlice
}

func deleteDiagnosticLogging(resourceId string, currentDiagSettings *[]string, meta interface{}) error {
	client := meta.(*ArmClient).monitorDiagnosticSettingsClient
	ctx := meta.(*ArmClient).StopContext

	for _, element := range *currentDiagSettings {
		_, err := client.Delete(ctx, resourceId, element)
		if err != nil {
			fmt.Println("lol")
		}
	}

	return nil
}

func expandMetricsConfiguration(metricsSettings *schema.Set) *[]insights.MetricSettings {
	returnMetricsSettings := make([]insights.MetricSettings, 0, metricsSettings.Len())

	for _, setting := range metricsSettings.List() {
		setting := setting.(map[string]interface{})
		category := setting["category"].(string)
		enabled := setting["enabled"].(bool)
		var retentionDays int32
		retentionDays = 4
		retentionPolicy := insights.RetentionPolicy{
			Days:    &retentionDays,
			Enabled: &enabled,
		}

		metricSetting := insights.MetricSettings{
			Category:        &category,
			Enabled:         &enabled,
			RetentionPolicy: &retentionPolicy,
		}
		returnMetricsSettings = append(returnMetricsSettings, metricSetting)
	}
	return &returnMetricsSettings
}

func expandLogConfiguration(logSettings *schema.Set) *[]insights.LogSettings {
	returnLogSettings := make([]insights.LogSettings, 0, logSettings.Len())

	for _, setting := range logSettings.List() {
		rawSetting := setting.(map[string]interface{})
		category := rawSetting["category"].(string)
		enabled := rawSetting["enabled"].(bool)
		var retentionDays int32
		retentionDays = 4
		retentionPolicy := insights.RetentionPolicy{
			Days:    &retentionDays,
			Enabled: &enabled,
		}

		logSetting := insights.LogSettings{
			Category:        &category,
			Enabled:         &enabled,
			RetentionPolicy: &retentionPolicy,
		}
		returnLogSettings = append(returnLogSettings, logSetting)
	}
	return &returnLogSettings
}

func expandDiagnosticSettings(settings map[string]interface{}) insights.DiagnosticSettings {
	diagnosticSettings := insights.DiagnosticSettings{}

	storageAccountID := settings["storage_account_id"].(string)
	eventHubName := settings["event_hub_name"].(string)
	eventHubAuthorizationRuleID := settings["event_hub_authorization_rule_id"].(string)
	workspaceID := settings["workspace_id"].(string)
	metricSettings := settings["metric_settings"]
	logSettings := settings["log_settings"]

	if metricSettings != nil {
		diagnosticSettings.Metrics = expandMetricsConfiguration(metricSettings.(*schema.Set))
	}

	if logSettings != nil {
		diagnosticSettings.Logs = expandLogConfiguration(logSettings.(*schema.Set))
	}

	if len(storageAccountID) > 0 {
		diagnosticSettings.StorageAccountID = &storageAccountID
	}

	if len(workspaceID) > 0 {
		diagnosticSettings.WorkspaceID = &workspaceID
	}

	if len(eventHubAuthorizationRuleID) > 0 && len(eventHubName) > 0 {
		diagnosticSettings.EventHubAuthorizationRuleID = &eventHubAuthorizationRuleID
		diagnosticSettings.EventHubName = &eventHubName
	}
	return diagnosticSettings
}
