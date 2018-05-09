package azurerm

import (
	"fmt"
	//"net/http"
	"testing"

	//"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMMonitorDiagnostics_basic(t *testing.T) {
	resourceName := "azurerm_monitor_diagnostics.test"
	objectName := "jbolltesting"
	config := testAccAzureRMMonitorDiagnostics_basic()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMMonitorDiagnosticsExists(resourceName, objectName),
				),
			},
		},
	})
}

func testCheckAzureRMMonitorDiagnosticsExists(name, objectName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		stateName := rs.Primary.Attributes["name"]
		if stateName != objectName {
			return fmt.Errorf("State inconsistent, %s does not match state name %s", stateName, objectName)
		}

		return nil
	}
}

func testAccAzureRMMonitorDiagnostics_basic() string {
	return fmt.Sprintf(`resource "azurerm_monitor_diagnostics" "test" {
    name = "jbolltesting"
	resource_id = "/subscriptions/xxx/resourceGroups/jbolltesting/providers/Microsoft.KeyVault/vaults/jbolltesting"
	storage_account_id = "/subscriptions/xxx/resourceGroups/jbolltesting/providers/Microsoft.Storage/storageAccounts/jbollvaulttestingaccount"
	metric_settings = {
		category = "AllMetrics"
		enabled = true
		retention_days = 2
	}
}`)
}
