// Code generated by go generate; DO NOT EDIT.
/***************************************************************
 *
 * Copyright (C) 2024, Pelican Project, Morgridge Institute for Research
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you
 * may not use this file except in compliance with the License.  You may
 * obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 ***************************************************************/

package param

import (
	"time"
)

type Config struct {
	Cache struct {
		AcceptedNamespaces []string
		Concurrency int
		DataLocation string
		EnableVoms bool
		ExportLocation string
		Port int
		RunLocation string
		Url string
		XRootDPrefix string
	}
	Client struct {
		DisableHttpProxy bool
		DisableProxyFallback bool
		MinimumDownloadSpeed int
		SlowTransferRampupTime int
		SlowTransferWindow int
		StoppedTransferTimeout int
		WorkerCount int
	}
	ConfigDir string
	Debug bool
	Director struct {
		AdvertisementTTL time.Duration
		CacheResponseHostnames []string
		DefaultResponse string
		EnableBroker bool
		GeoIPLocation string
		MaxMindKeyFile string
		MaxStatResponse int
		MinStatResponse int
		OriginCacheHealthTestInterval time.Duration
		OriginResponseHostnames []string
		StatConcurrencyLimit int
		StatTimeout time.Duration
	}
	DisableHttpProxy bool
	DisableProxyFallback bool
	Federation struct {
		BrokerUrl string
		DirectorUrl string
		DiscoveryUrl string
		JwkUrl string
		NamespaceUrl string
		RegistryUrl string
		TopologyNamespaceUrl string
		TopologyReloadInterval time.Duration
		TopologyUrl string
	}
	GeoIPOverrides interface{}
	Issuer struct {
		AuthenticationSource string
		AuthorizationTemplates interface{}
		GroupFile string
		GroupRequirements []string
		GroupSource string
		OIDCAuthenticationRequirements interface{}
		OIDCAuthenticationUserClaim string
		QDLLocation string
		ScitokensServerLocation string
		TomcatLocation string
	}
	IssuerKey string
	Logging struct {
		Cache struct {
			Ofs string
			Pss string
			Scitokens string
			Xrd string
		}
		DisableProgressBars bool
		Level string
		LogLocation string
		Origin struct {
			Cms string
			Pfc string
			Pss string
			Scitokens string
			Xrootd string
		}
	}
	MinimumDownloadSpeed int
	Monitoring struct {
		AggregatePrefixes []string
		DataLocation string
		MetricAuthorization bool
		PortHigher int
		PortLower int
		PromQLAuthorization bool
		TokenExpiresIn time.Duration
		TokenRefreshInterval time.Duration
	}
	OIDC struct {
		AuthorizationEndpoint string
		ClientID string
		ClientIDFile string
		ClientRedirectHostname string
		ClientSecretFile string
		DeviceAuthEndpoint string
		Issuer string
		TokenEndpoint string
		UserInfoEndpoint string
	}
	Origin struct {
		EnableBroker bool
		EnableCmsd bool
		EnableDirListing bool
		EnableFallbackRead bool
		EnableIssuer bool
		EnablePublicReads bool
		EnableUI bool
		EnableVoms bool
		EnableWrite bool
		ExportVolume string
		Mode string
		Multiuser bool
		NamespacePrefix string
		Port int
		RunLocation string
		S3AccessKeyfile string
		S3Bucket string
		S3Region string
		S3SecretKeyfile string
		S3ServiceName string
		S3ServiceUrl string
		S3UrlStyle string
		ScitokensDefaultUser string
		ScitokensMapSubject bool
		ScitokensNameMapFile string
		ScitokensRestrictedPaths []string
		ScitokensUsernameClaim string
		SelfTest bool
		SelfTestInterval time.Duration
		Url string
		XRootDPrefix string
	}
	Plugin struct {
		Token string
	}
	Registry struct {
		AdminUsers []string
		CustomRegistrationFields interface{}
		DbLocation string
		Institutions interface{}
		InstitutionsUrl string
		InstitutionsUrlReloadMinutes time.Duration
		RequireCacheApproval bool
		RequireKeyChaining bool
		RequireOriginApproval bool
	}
	Server struct {
		EnableUI bool
		ExternalWebUrl string
		Hostname string
		IssuerHostname string
		IssuerJwks string
		IssuerPort int
		IssuerUrl string
		Modules []string
		RegistrationRetryInterval time.Duration
		SessionSecretFile string
		TLSCACertificateDirectory string
		TLSCACertificateFile string
		TLSCAKey string
		TLSCertificate string
		TLSKey string
		UIActivationCodeFile string
		UILoginRateLimit int
		UIPasswordFile string
		WebConfigFile string
		WebHost string
		WebPort int
	}
	Shoveler struct {
		AMQPExchange string
		AMQPTokenLocation string
		Enable bool
		IPMapping interface{}
		MessageQueueProtocol string
		OutputDestinations []string
		PortHigher int
		PortLower int
		QueueDirectory string
		StompCert string
		StompCertKey string
		StompPassword string
		StompUsername string
		Topic string
		URL string
		VerifyHeader bool
	}
	StagePlugin struct {
		Hook bool
		MountPrefix string
		OriginPrefix string
		ShadowOriginPrefix string
	}
	TLSSkipVerify bool
	Transport struct {
		DialerKeepAlive time.Duration
		DialerTimeout time.Duration
		ExpectContinueTimeout time.Duration
		IdleConnTimeout time.Duration
		MaxIdleConns int
		ResponseHeaderTimeout time.Duration
		TLSHandshakeTimeout time.Duration
	}
	Xrootd struct {
		Authfile string
		DetailedMonitoringHost string
		LocalMonitoringHost string
		MacaroonsKeyFile string
		ManagerHost string
		Mount string
		Port int
		RobotsTxtFile string
		RunLocation string
		ScitokensConfig string
		Sitename string
		SummaryMonitoringHost string
	}
}


type configWithType struct {
	Cache struct {
		AcceptedNamespaces struct { Type string; Value []string }
		Concurrency struct { Type string; Value int }
		DataLocation struct { Type string; Value string }
		EnableVoms struct { Type string; Value bool }
		ExportLocation struct { Type string; Value string }
		Port struct { Type string; Value int }
		RunLocation struct { Type string; Value string }
		Url struct { Type string; Value string }
		XRootDPrefix struct { Type string; Value string }
	}
	Client struct {
		DisableHttpProxy struct { Type string; Value bool }
		DisableProxyFallback struct { Type string; Value bool }
		MinimumDownloadSpeed struct { Type string; Value int }
		SlowTransferRampupTime struct { Type string; Value int }
		SlowTransferWindow struct { Type string; Value int }
		StoppedTransferTimeout struct { Type string; Value int }
		WorkerCount struct { Type string; Value int }
	}
	ConfigDir struct { Type string; Value string }
	Debug struct { Type string; Value bool }
	Director struct {
		AdvertisementTTL struct { Type string; Value time.Duration }
		CacheResponseHostnames struct { Type string; Value []string }
		DefaultResponse struct { Type string; Value string }
		EnableBroker struct { Type string; Value bool }
		GeoIPLocation struct { Type string; Value string }
		MaxMindKeyFile struct { Type string; Value string }
		MaxStatResponse struct { Type string; Value int }
		MinStatResponse struct { Type string; Value int }
		OriginCacheHealthTestInterval struct { Type string; Value time.Duration }
		OriginResponseHostnames struct { Type string; Value []string }
		StatConcurrencyLimit struct { Type string; Value int }
		StatTimeout struct { Type string; Value time.Duration }
	}
	DisableHttpProxy struct { Type string; Value bool }
	DisableProxyFallback struct { Type string; Value bool }
	Federation struct {
		BrokerUrl struct { Type string; Value string }
		DirectorUrl struct { Type string; Value string }
		DiscoveryUrl struct { Type string; Value string }
		JwkUrl struct { Type string; Value string }
		NamespaceUrl struct { Type string; Value string }
		RegistryUrl struct { Type string; Value string }
		TopologyNamespaceUrl struct { Type string; Value string }
		TopologyReloadInterval struct { Type string; Value time.Duration }
		TopologyUrl struct { Type string; Value string }
	}
	GeoIPOverrides struct { Type string; Value interface{} }
	Issuer struct {
		AuthenticationSource struct { Type string; Value string }
		AuthorizationTemplates struct { Type string; Value interface{} }
		GroupFile struct { Type string; Value string }
		GroupRequirements struct { Type string; Value []string }
		GroupSource struct { Type string; Value string }
		OIDCAuthenticationRequirements struct { Type string; Value interface{} }
		OIDCAuthenticationUserClaim struct { Type string; Value string }
		QDLLocation struct { Type string; Value string }
		ScitokensServerLocation struct { Type string; Value string }
		TomcatLocation struct { Type string; Value string }
	}
	IssuerKey struct { Type string; Value string }
	Logging struct {
		Cache struct {
			Ofs struct { Type string; Value string }
			Pss struct { Type string; Value string }
			Scitokens struct { Type string; Value string }
			Xrd struct { Type string; Value string }
		}
		DisableProgressBars struct { Type string; Value bool }
		Level struct { Type string; Value string }
		LogLocation struct { Type string; Value string }
		Origin struct {
			Cms struct { Type string; Value string }
			Pfc struct { Type string; Value string }
			Pss struct { Type string; Value string }
			Scitokens struct { Type string; Value string }
			Xrootd struct { Type string; Value string }
		}
	}
	MinimumDownloadSpeed struct { Type string; Value int }
	Monitoring struct {
		AggregatePrefixes struct { Type string; Value []string }
		DataLocation struct { Type string; Value string }
		MetricAuthorization struct { Type string; Value bool }
		PortHigher struct { Type string; Value int }
		PortLower struct { Type string; Value int }
		PromQLAuthorization struct { Type string; Value bool }
		TokenExpiresIn struct { Type string; Value time.Duration }
		TokenRefreshInterval struct { Type string; Value time.Duration }
	}
	OIDC struct {
		AuthorizationEndpoint struct { Type string; Value string }
		ClientID struct { Type string; Value string }
		ClientIDFile struct { Type string; Value string }
		ClientRedirectHostname struct { Type string; Value string }
		ClientSecretFile struct { Type string; Value string }
		DeviceAuthEndpoint struct { Type string; Value string }
		Issuer struct { Type string; Value string }
		TokenEndpoint struct { Type string; Value string }
		UserInfoEndpoint struct { Type string; Value string }
	}
	Origin struct {
		EnableBroker struct { Type string; Value bool }
		EnableCmsd struct { Type string; Value bool }
		EnableDirListing struct { Type string; Value bool }
		EnableFallbackRead struct { Type string; Value bool }
		EnableIssuer struct { Type string; Value bool }
		EnablePublicReads struct { Type string; Value bool }
		EnableUI struct { Type string; Value bool }
		EnableVoms struct { Type string; Value bool }
		EnableWrite struct { Type string; Value bool }
		ExportVolume struct { Type string; Value string }
		Mode struct { Type string; Value string }
		Multiuser struct { Type string; Value bool }
		NamespacePrefix struct { Type string; Value string }
		Port struct { Type string; Value int }
		RunLocation struct { Type string; Value string }
		S3AccessKeyfile struct { Type string; Value string }
		S3Bucket struct { Type string; Value string }
		S3Region struct { Type string; Value string }
		S3SecretKeyfile struct { Type string; Value string }
		S3ServiceName struct { Type string; Value string }
		S3ServiceUrl struct { Type string; Value string }
		S3UrlStyle struct { Type string; Value string }
		ScitokensDefaultUser struct { Type string; Value string }
		ScitokensMapSubject struct { Type string; Value bool }
		ScitokensNameMapFile struct { Type string; Value string }
		ScitokensRestrictedPaths struct { Type string; Value []string }
		ScitokensUsernameClaim struct { Type string; Value string }
		SelfTest struct { Type string; Value bool }
		SelfTestInterval struct { Type string; Value time.Duration }
		Url struct { Type string; Value string }
		XRootDPrefix struct { Type string; Value string }
	}
	Plugin struct {
		Token struct { Type string; Value string }
	}
	Registry struct {
		AdminUsers struct { Type string; Value []string }
		CustomRegistrationFields struct { Type string; Value interface{} }
		DbLocation struct { Type string; Value string }
		Institutions struct { Type string; Value interface{} }
		InstitutionsUrl struct { Type string; Value string }
		InstitutionsUrlReloadMinutes struct { Type string; Value time.Duration }
		RequireCacheApproval struct { Type string; Value bool }
		RequireKeyChaining struct { Type string; Value bool }
		RequireOriginApproval struct { Type string; Value bool }
	}
	Server struct {
		EnableUI struct { Type string; Value bool }
		ExternalWebUrl struct { Type string; Value string }
		Hostname struct { Type string; Value string }
		IssuerHostname struct { Type string; Value string }
		IssuerJwks struct { Type string; Value string }
		IssuerPort struct { Type string; Value int }
		IssuerUrl struct { Type string; Value string }
		Modules struct { Type string; Value []string }
		RegistrationRetryInterval struct { Type string; Value time.Duration }
		SessionSecretFile struct { Type string; Value string }
		TLSCACertificateDirectory struct { Type string; Value string }
		TLSCACertificateFile struct { Type string; Value string }
		TLSCAKey struct { Type string; Value string }
		TLSCertificate struct { Type string; Value string }
		TLSKey struct { Type string; Value string }
		UIActivationCodeFile struct { Type string; Value string }
		UILoginRateLimit struct { Type string; Value int }
		UIPasswordFile struct { Type string; Value string }
		WebConfigFile struct { Type string; Value string }
		WebHost struct { Type string; Value string }
		WebPort struct { Type string; Value int }
	}
	Shoveler struct {
		AMQPExchange struct { Type string; Value string }
		AMQPTokenLocation struct { Type string; Value string }
		Enable struct { Type string; Value bool }
		IPMapping struct { Type string; Value interface{} }
		MessageQueueProtocol struct { Type string; Value string }
		OutputDestinations struct { Type string; Value []string }
		PortHigher struct { Type string; Value int }
		PortLower struct { Type string; Value int }
		QueueDirectory struct { Type string; Value string }
		StompCert struct { Type string; Value string }
		StompCertKey struct { Type string; Value string }
		StompPassword struct { Type string; Value string }
		StompUsername struct { Type string; Value string }
		Topic struct { Type string; Value string }
		URL struct { Type string; Value string }
		VerifyHeader struct { Type string; Value bool }
	}
	StagePlugin struct {
		Hook struct { Type string; Value bool }
		MountPrefix struct { Type string; Value string }
		OriginPrefix struct { Type string; Value string }
		ShadowOriginPrefix struct { Type string; Value string }
	}
	TLSSkipVerify struct { Type string; Value bool }
	Transport struct {
		DialerKeepAlive struct { Type string; Value time.Duration }
		DialerTimeout struct { Type string; Value time.Duration }
		ExpectContinueTimeout struct { Type string; Value time.Duration }
		IdleConnTimeout struct { Type string; Value time.Duration }
		MaxIdleConns struct { Type string; Value int }
		ResponseHeaderTimeout struct { Type string; Value time.Duration }
		TLSHandshakeTimeout struct { Type string; Value time.Duration }
	}
	Xrootd struct {
		Authfile struct { Type string; Value string }
		DetailedMonitoringHost struct { Type string; Value string }
		LocalMonitoringHost struct { Type string; Value string }
		MacaroonsKeyFile struct { Type string; Value string }
		ManagerHost struct { Type string; Value string }
		Mount struct { Type string; Value string }
		Port struct { Type string; Value int }
		RobotsTxtFile struct { Type string; Value string }
		RunLocation struct { Type string; Value string }
		ScitokensConfig struct { Type string; Value string }
		Sitename struct { Type string; Value string }
		SummaryMonitoringHost struct { Type string; Value string }
	}
}
