package tozny

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tozny/e3db-clients-go/identityClient"
)

// resourcePAMJiraPlugin returns the schema and methods for configuring PAM Jira Plugin integrations
func resourcePAMJiraPlugin() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePAMJiraPluginCreate,
		ReadContext:   resourcePAMJiraPluginRead,
		DeleteContext: resourcePAMJiraPluginDelete,
		UpdateContext: resourcePAMJiraPluginUpdate,
		Schema: map[string]*schema.Schema{
			"client_credentials_filepath": {
				Description:   "The filepath to Tozny client credentials for the Terraform provider to use when provisioning this realm provider.",
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				ForceNew:      true,
				ConflictsWith: []string{"client_credentials_config"},
			},
			"client_credentials_config": {
				Description:   "The Tozny account client configuration as a JSON string",
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"client_credentials_filepath"},
			},
			"realm_name": {
				Description: "Server defined unique identifier for a realm",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"jira_host_url": {
				Description: "The url of the jira instance with no protocol or trailing slash",
				Type:        schema.TypeString,
				Required:    true,
			},
			"jira_bot_user_email": {
				Description: "The email of the Jira user that performs actions on behalf of TozID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"jira_bot_user_api_key": {
				Description: "The api key of the Jira user that performs actions on behalf of TozID. Ideally, this value should come from a secret store.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"automation_auth_header": {
				Description: "The authentication header to be used in requests from Jira using this integration",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourcePAMJiraPluginCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	pluginParams := identityClient.CreatePAMJiraPluginRequest{
		RealmName:     d.Get("realm_name").(string),
		JiraHostURL:   d.Get("jira_host_url").(string),
		BotUserEmail:  d.Get("jira_bot_user_email").(string),
		BotUserAPIKey: d.Get("jira_bot_user_api_key").(string),
	}
	plugin, err := toznySDK.CreatePAMJiraPlugin(ctx, pluginParams)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create plugin: %s", err))
	}

	d.Set("automation_auth_header", plugin.AutomationAuthHeader)
	d.SetId(strconv.FormatInt(plugin.ID, 10))

	return diags
}

func resourcePAMJiraPluginRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to parse plugin id: %s %s", d.Id(), err))
	}

	plugin, err := toznySDK.DescribePAMJiraPlugin(ctx, identityClient.DescribePAMJiraPluginRequest{PluginID: id})
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read plugin: %s", err))
	}

	d.Set("automation_auth_header", plugin.AutomationAuthHeader)
	d.Set("jira_host_url", plugin.JiraHostURL)
	d.Set("jira_bot_user_email", plugin.BotUserEmail)

	return diags
}

func resourcePAMJiraPluginDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to parse plugin id: %s %s", d.Id(), err))
	}

	err = toznySDK.DeletePAMJiraPlugin(ctx, identityClient.DeletePAMJiraPluginRequest{PluginID: id})
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete plugin: %s", err))
	}

	d.SetId("")

	return diags
}
func resourcePAMJiraPluginUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	toznySDK, err := MakeToznySDK(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	pluginID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to parse plugin id: %s %s", d.Id(), err))
	}

	// Check if relevant properties have changed
	if d.HasChanges("jira_host_url", "jira_bot_user_email", "jira_bot_user_api_key") {
		// Update the plugin!
		updateParams := identityClient.UpdatePAMJiraPluginRequest{
			PluginID:      pluginID,
			JiraHostURL:   d.Get("jira_host_url").(string),
			BotUserEmail:  d.Get("jira_bot_user_email").(string),
			BotUserAPIKey: d.Get("jira_bot_user_api_key").(string),
		}
		_, err := toznySDK.UpdatePAMJiraPlugin(ctx, updateParams)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to update plugin: %s", err))
		}
	}

	return diags
}
