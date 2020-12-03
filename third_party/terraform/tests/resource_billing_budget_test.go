package google

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBillingBudget_billingBudgetBasicExample(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"billing_acct":  getTestBillingAccountFromEnv(t),
		"random_suffix": randString(t, 10),
	}

	vcrTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {},
		},
		CheckDestroy: testAccCheckBillingBudgetDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingBudget_billingBudgetBasicExample(context),
			},
		},
	})
}

func testAccBillingBudget_billingBudgetBasicExample(context map[string]interface{}) string {
	return Nprintf(`
data "google_billing_account" "account" {
  billing_account = "%{billing_acct}"
}

resource "google_billing_budget" "budget" {
  billing_account = data.google_billing_account.account.id
  display_name = "Example Billing Budget%{random_suffix}"
  amount {
    specified_amount {
      currency_code = "USD"
      units = "100000"
    }
  }
  threshold_rules {
      threshold_percent =  0.5
  }
}
`, context)
}

func TestAccBillingBudget_billingBudgetLastperiodExample(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"billing_acct":  getTestBillingAccountFromEnv(t),
		"project":       getTestProjectFromEnv(),
		"random_suffix": randString(t, 10),
	}

	vcrTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {},
		},
		CheckDestroy: testAccCheckBillingBudgetDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingBudget_billingBudgetLastperiodExample(context),
			},
		},
	})
}

func testAccBillingBudget_billingBudgetLastperiodExample(context map[string]interface{}) string {
	return Nprintf(`
data "google_billing_account" "account" {
  billing_account = "%{billing_acct}"
}

data "google_project" "project" {
}

resource "google_billing_budget" "budget" {
  billing_account = data.google_billing_account.account.id
  display_name = "Example Billing Budget%{random_suffix}"
  
  budget_filter {
    projects = ["projects/${data.google_project.project.number}"]
  }

  amount {
    last_period_amount = true
  }

  threshold_rules {
      threshold_percent =  10.0
      # Typically threshold_percent would be set closer to 1.0 (100%).
      # It has been purposely set high (10.0 / 1000%) in this example
      # so it does not trigger alerts during automated testing.
  }
}
`, context)
}

func TestAccBillingBudget_billingBudgetFilterExample(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"billing_acct":  getTestBillingAccountFromEnv(t),
		"project":       getTestProjectFromEnv(),
		"random_suffix": randString(t, 10),
	}

	vcrTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {},
		},
		CheckDestroy: testAccCheckBillingBudgetDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingBudget_billingBudgetFilterExample(context),
			},
		},
	})
}

func testAccBillingBudget_billingBudgetFilterExample(context map[string]interface{}) string {
	return Nprintf(`
data "google_billing_account" "account" {
  billing_account = "%{billing_acct}"
}

data "google_project" "project" {
}

resource "google_billing_budget" "budget" {
  billing_account = data.google_billing_account.account.id
  display_name = "Example Billing Budget%{random_suffix}"

  budget_filter {
    projects = ["projects/${data.google_project.project.number}"]
    credit_types_treatment = "EXCLUDE_ALL_CREDITS"
    services = ["services/24E6-581D-38E5"] # Bigquery
  }

  amount {
    specified_amount {
      currency_code = "USD"
      units = "100000"
    }
  }

  threshold_rules {
    threshold_percent = 0.5
  }
  threshold_rules {
    threshold_percent = 0.9
    spend_basis = "FORECASTED_SPEND"
  }
}
`, context)
}

func TestAccBillingBudget_billingBudgetNotifyExample(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"billing_acct":  getTestBillingAccountFromEnv(t),
		"project":       getTestProjectFromEnv(),
		"random_suffix": randString(t, 10),
	}

	vcrTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {},
		},
		CheckDestroy: testAccCheckBillingBudgetDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingBudget_billingBudgetNotifyExample(context),
			},
		},
	})
}

func testAccBillingBudget_billingBudgetNotifyExample(context map[string]interface{}) string {
	return Nprintf(`
data "google_billing_account" "account" {
  billing_account = "%{billing_acct}"
}

data "google_project" "project" {
}

resource "google_billing_budget" "budget" {
  billing_account = data.google_billing_account.account.id
  display_name    = "Example Billing Budget%{random_suffix}"

  budget_filter {
    projects = ["projects/${data.google_project.project.number}"]
  }

  amount {
    specified_amount {
      currency_code = "USD"
      units         = "100000"
    }
  }

  threshold_rules {
    threshold_percent = 1.0
  }
  threshold_rules {
    threshold_percent = 1.0
    spend_basis       = "FORECASTED_SPEND"
  }
  
  all_updates_rule {
    monitoring_notification_channels = [
      google_monitoring_notification_channel.notification_channel.id,
    ]
    disable_default_iam_recipients = true
  }
}

resource "google_monitoring_notification_channel" "notification_channel" {
  display_name = "Example Notification Channel%{random_suffix}"
  type         = "email"
  
  labels = {
    email_address = "address@example.com"
  }
}
`, context)
}

func testAccCheckBillingBudgetDestroyProducer(t *testing.T) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "google_billing_budget" {
				continue
			}
			if strings.HasPrefix(name, "data.") {
				continue
			}

			config := googleProviderConfig(t)

			url, err := replaceVarsForTest(config, rs, "{{BillingBasePath}}{{name}}")
			if err != nil {
				return err
			}

			billingProject := ""

			if config.BillingProject != "" {
				billingProject = config.BillingProject
			}

			_, err = sendRequest(config, "GET", billingProject, url, config.userAgent, nil)
			if err == nil {
				return fmt.Errorf("BillingBudget still exists at %s", url)
			}
		}

		return nil
	}
}
