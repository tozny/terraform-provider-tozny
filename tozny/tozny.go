package tozny

import (
  "context"
  "encoding/json"
  "io/ioutil"
  "os"

  "github.com/tozny/e3db-clients-go"
  "github.com/tozny/e3db-clients-go/identityClient"
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

// ToznyBrokerIdentityConfig wraps values for creating a Tozny Identity for brokering Realm activities
type ToznyBrokerIdentityConfig struct {
  ClientRegistrationToken string
  Name                    string
  RealmName               string
}

// SecretKeys wraps private key material from keypair(s) used for encryption and signing operations
type SecretKeys struct {
  PrivateEncryptionKey e3dbClients.Key
  PrivateSigningKey    e3dbClients.Key
}

// MakeToznyBrokerIdentity generates the configuration necessary to create a brokering Identity, returning the
// public Identity configuration and secret key material separately and error (if any).
func MakeToznyBrokerIdentity(config ToznyBrokerIdentityConfig) (identityClient.Identity, SecretKeys, error) {
  var broker identityClient.Identity
  var secretKeys SecretKeys

  signingKeyPair, err := e3dbClients.GenerateSigningKeys()

  if err != nil {
    return broker, secretKeys, err
  }

  encryptionKeyPair, err := e3dbClients.GenerateKeyPair()

  if err != nil {
    return broker, secretKeys, err
  }

  broker = identityClient.Identity{
    Name:        config.Name,
    PublicKeys:  map[string]string{encryptionKeyPair.Public.Type: encryptionKeyPair.Public.Material},
    SigningKeys: map[string]string{signingKeyPair.Public.Type: signingKeyPair.Public.Material},
  }

  secretKeys.PrivateEncryptionKey, secretKeys.PrivateSigningKey = encryptionKeyPair.Private, signingKeyPair.Private

  return broker, secretKeys, nil
}

// SaveToznyBrokerIdentity serialize a broker identity to the specified file,
// returning error (if any).
func SaveToznyBrokerIdentity(filepath string, broker identityClient.Identity) error {
  file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0400)

  if err != nil {
    return err
  }

  defer file.Close()

  if err = json.NewEncoder(file).Encode(&broker); err != nil {
    return err
  }

  return nil
}

// LoadToznyBrokerIdentity loads a serialized broker identity from file into
// the provided response struct, returning error (if any).
func LoadToznyBrokerIdentity(filepath string, broker *identityClient.Identity) error {
  bytes, err := ioutil.ReadFile(filepath)

  if err != nil {
    return err
  }

  err = json.Unmarshal(bytes, broker)

  if err != nil {
    return err
  }

  return nil
}
