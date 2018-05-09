package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/monitor/mgmt/2017-05-01-preview/insights"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmMonitorDiagnostics() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmMonitorDiagnosticsCreate,
		Read:   resourceArmMonitorDiagnosticsRead,
		Update: resourceArmMonitorDiagnosticsCreate,
		Delete: resourceArmMonitorDiagnosticsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": {
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
	}
}

func resourceArmMonitorDiagnosticsCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).monitorDiagnosticSettingsClient
	ctx := meta.(*ArmClient).StopContext
	log.Printf("[INFO] preparing arguments for Azure ARM KeyVault creation.")

	name := d.Get("name").(string)
	resource_id := d.Get("resource_id").(string)
	storage_account_id := d.Get("storage_account_id").(string)
	event_hub_name := d.Get("event_hub_name").(string)
	event_hub_authorization_rule_id := d.Get("event_hub_authorization_rule_id").(string)
	workspace_id := d.Get("workspace_id").(string)
	metric_settings := d.Get("metric_settings")
	log_settings := d.Get("log_settings")

	diagnosticSettings := insights.DiagnosticSettings{}

	if metric_settings != nil {
		diagnosticSettings.Metrics = expandMetricsConfiguration(metric_settings.(*schema.Set))
	}

	if log_settings != nil {
		diagnosticSettings.Logs = expandLogConfiguration(log_settings.(*schema.Set))
	}

	if len(storage_account_id) > 0 {
		diagnosticSettings.StorageAccountID = &storage_account_id
	}

	if len(workspace_id) > 0 {
		diagnosticSettings.WorkspaceID = &workspace_id
	}

	if len(event_hub_authorization_rule_id) > 0 && len(event_hub_name) > 0 {
		diagnosticSettings.EventHubAuthorizationRuleID = &event_hub_authorization_rule_id
		diagnosticSettings.EventHubName = &event_hub_name
	}

	_, err := client.CreateOrUpdate(
		ctx,
		resource_id,
		insights.DiagnosticSettingsResource{
			Name:               &name,
			DiagnosticSettings: &diagnosticSettings,
		},
		name)
	if err != nil {
		return err
	}

	read, err := client.Get(ctx, resource_id, name)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read Diagnostic Settings")
	}

	d.SetId(*read.ID)

	return resourceArmMonitorDiagnosticsRead(d, meta)
}

func resourceArmMonitorDiagnosticsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).monitorDiagnosticSettingsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	resource_id := d.Get("resource_id").(string)

	resp, err := client.Get(ctx, resource_id, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure KeyVault %s: %+v", name, err)
	}

	d.SetId(*resp.ID)

	return nil
}

func resourceArmMonitorDiagnosticsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).monitorDiagnosticSettingsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	resource_id := d.Get("resource_id").(string)

	_, err := client.Delete(ctx, resource_id, name)

	return err
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
