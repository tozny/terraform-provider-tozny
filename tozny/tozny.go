package tozny

import (
  "context"
  "github.com/tozny/e3db-go/v2"
)

// MakeToznySession uses Terraform provider and resource configuration to create a Tozny session
// for communicating to account and client level APIs, returning an SDK and Account session (with API token) and error
// (if any).
func MakeToznySession(ctx context.Context, sdkCredentialsFilePath string, terraformProviderConfig interface{}) (*e3db.ToznySDKV3, e3db.Account, error) {
  var account e3db.Account

  toznySDK, err := MakeToznySDK(sdkCredentialsFilePath, terraformProviderConfig)

  if err != nil {
    return toznySDK, account, err
  }

  account, err = toznySDK.Login(ctx, toznySDK.AccountUsername, toznySDK.AccountPassword, "password", toznySDK.APIEndpoint)

  if err != nil {
    return toznySDK, account, err
  }

  return toznySDK, account, nil
}

// MakeToznySDK uses Terraform provider and resource configuration to create a Tozny SDK provider,
// returning the SDK and error (if any).
func MakeToznySDK(sdkCredentialsFilePath string, terraformProviderConfig interface{}) (*e3db.ToznySDKV3, error) {
  toznySDK := terraformProviderConfig.(*e3db.ToznySDKV3)

  var err error

  if sdkCredentialsFilePath != "" {
    toznySDK, err = e3db.GetSDKV3(sdkCredentialsFilePath)

    if err != nil {
      return toznySDK, err
    }

  }

  return toznySDK, nil
}
