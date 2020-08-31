package tozny

import (
  "context"
  "fmt"

  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/tozny/e3db-clients-go/identityClient"
  "github.com/tozny/e3db-clients-go/pdsClient"
)

// resourceRealmBrokerDelegation returns the schema and methods for provisioning a Tozny Realm Broker Delegation
func resourceRealmBrokerDelegation() *schema.Resource {
  return &schema.Resource{
    CreateContext: resourceRealmBrokerDelegationCreate,
    ReadContext:   resourceRealmBrokerDelegationRead,
    DeleteContext: resourceRealmBrokerDelegationDelete,
    Schema: map[string]*schema.Schema{
      "realm_broker_identity_credentials_filepath": {
        Description: "The filepath to load the realm broker identity to delegate access to.",
        Type:        schema.TypeString,
        Required:    true,
        ForceNew:    true,
      },
      "client_credentials_filepath": {
        Description: "The filepath to Tozny client credentials for the provider to use when provisioning this broker delegation.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "use_tozny_hosted_broker": {
        Description: "Whether to delegate realm brokering to the Tozny Hosted Broker. Defaults to true.",
        Type:        schema.TypeBool,
        Optional:    true,
        Default:     true,
        ForceNew:    true,
      },
      "client_id_to_delegate_brokering": {
        Description: "Client ID to delegate realm brokering to.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "delegated_broker_client_id": {
        Description: "The ID of the client realm brokering is delegated to.",
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
      },
      "broker_token_record_id": {
        Description: "ID of the  TozStore record containing material to derive the realm broker identity credentials.",
        Type:        schema.TypeString,
        Computed:    true,
        ForceNew:    true,
      },
    },
  }
}

func resourceRealmBrokerDelegationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  var diags diag.Diagnostics

  toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

  toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

  if err != nil {
    return diag.FromErr(err)
  }

  var broker identityClient.Identity

  err = LoadToznyBrokerIdentity(d.Get("realm_broker_identity_credentials_filepath").(string), &broker)

  if err != nil {
    return diag.FromErr(err)
  }

  brokerNoteToken, err := toznySDK.GenerateRealmBrokerNoteToken(ctx, broker)

  if err != nil {
    return diag.FromErr(err)
  }

  brokerNoteTokenRecordType := fmt.Sprintf("%s.backup.token", broker.RealmName)

  sdkClientID := toznySDK.E3dbPDSClient.ClientID

  brokerNoteTokenRecordToWrite := pdsClient.Record{
    Data: map[string]string{"token": brokerNoteToken},
    Metadata: pdsClient.Meta{
      Type:     brokerNoteTokenRecordType,
      WriterID: sdkClientID,
      UserID:   sdkClientID,
    },
  }

  encryptedBrokerNoteTokenRecordToWrite, err := toznySDK.EncryptRecord(ctx, brokerNoteTokenRecordToWrite)

  record, err := toznySDK.E3dbPDSClient.WriteRecord(ctx, encryptedBrokerNoteTokenRecordToWrite)

  if err != nil {
    return diag.FromErr(err)
  }

  brokerNoteTokenRecordID := record.Metadata.RecordID

  d.Set("broker_token_record_id", brokerNoteTokenRecordID)

  clientIDToDelegateBrokering := d.Get("delegated_broker_client_id").(string)

  if d.Get("use_tozny_hosted_broker").(bool) {
    toznyHostedBrokerInfo, err := toznySDK.GetToznyHostedBrokerInfo(ctx)

    if err != nil {
      return diag.FromErr(err)
    }

    clientIDToDelegateBrokering = toznyHostedBrokerInfo.ClientID.String()
  }

  if clientIDToDelegateBrokering != "" {
    err := toznySDK.AddAuthorizedSharer(ctx, pdsClient.AddAuthorizedWriterRequest{
      UserID:       sdkClientID,
      WriterID:     sdkClientID,
      AuthorizerID: clientIDToDelegateBrokering,
      RecordType:   brokerNoteTokenRecordType,
    })

    if err != nil {
      return diag.FromErr(err)
    }

    d.Set("delegated_broker_client_id", clientIDToDelegateBrokering)
  }

  // Associate created realm broker identity with Terraform state and signal success
  d.SetId(brokerNoteTokenRecordID)

  return diags
}

func resourceRealmBrokerDelegationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  var diags diag.Diagnostics

  toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

  toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

  if err != nil {
    return diag.FromErr(err)
  }

  batchRecords, err := toznySDK.BatchGetRecords(ctx, pdsClient.BatchGetRecordsRequest{
    RecordIDs:   []string{d.Get("broker_token_record_id").(string)},
    IncludeData: false,
  })

  if err != nil {
    return diag.FromErr(err)
  }

  if len(batchRecords.Records) == 0 {
    d.Set("broker_token_record_id", "")
  }

  return diags
}

func resourceRealmBrokerDelegationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
  var diags diag.Diagnostics

  toznyClientCredentialsFilePath := d.Get("client_credentials_filepath").(string)

  toznySDK, err := MakeToznySDK(toznyClientCredentialsFilePath, m)

  if err != nil {
    return diag.FromErr(err)
  }

  var broker identityClient.Identity

  err = LoadToznyBrokerIdentity(d.Get("realm_broker_identity_credentials_filepath").(string), &broker)

  if err != nil {
    return diag.FromErr(err)
  }

  delegatedBrokerClientID := d.Get("delegated_broker_client_id").(string)

  sdkClientID := toznySDK.E3dbPDSClient.ClientID

  brokerNoteTokenRecordType := fmt.Sprintf("%s.backup.token", broker.RealmName)

  if delegatedBrokerClientID != "" {
    err := toznySDK.RemoveAuthorizedSharer(ctx, pdsClient.AddAuthorizedWriterRequest{
      UserID:       sdkClientID,
      WriterID:     sdkClientID,
      AuthorizerID: delegatedBrokerClientID,
      RecordType:   brokerNoteTokenRecordType,
    })

    if err != nil {
      return diag.FromErr(err)
    }
  }

  err = toznySDK.DeleteRecord(ctx, pdsClient.DeleteRecordRequest{
    RecordID: d.Get("broker_token_record_id").(string),
  })

  if err != nil {
    return diag.FromErr(err)
  }

  // Delete from Terraform state and signal success
  d.SetId("")

  return diags
}
