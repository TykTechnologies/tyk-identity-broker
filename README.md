Table of Contents
=================

   * [Tyk Identity Broker (TIB)](#tyk-identity-broker-tib)
      * [What is Tyk Identity Broker?](#what-is-tyk-identity-broker)
         * [Requirements and dependencies](#requirements-and-dependencies)
         * [Installation](#installation)
            * [Run via docker](#run-via-docker)
         * [Usage](#usage)
         * [How it works](#how-it-works)
            * [Identity Providers](#identity-providers)
            * [Identity Handlers](#identity-handlers)
      * [How to configure TIB](#how-to-configure-tib)
         * [The tib.conf file](#the-tibconf-file)
            * [Secret](#secret)
            * [HttpServerOptions.UseSSL](#httpserveroptionsusessl)
            * [HttpServerOptions.CertFile](#httpserveroptionscertfile)
            * [HttpServerOptions.KeyFile](#httpserveroptionskeyfile)
            * [SSLInsecureSkipVerify](#sslinsecureskipverify)
            * [BackEnd](#backend)
            * [BackEnd.Hosts](#backendhosts)
            * [BackEnd.Password](#backendpassword)
            * [BackEnd.Database](#backenddatabase)
            * [BackEnd.EnableCluster](#backendenablecluster)
            * [BackEnd.MaxIdle](#backendmaxidle)
            * [BackEnd.MaxActive](#backendmaxactive)
            * [TykAPISettings](#tykapisettings)
            * [TykAPISettings.GatewayConfig.Endpoint](#tykapisettingsgatewayconfigendpoint)
            * [TykAPISettings.GatewayConfig.Port](#tykapisettingsgatewayconfigport)
            * [TykAPISettings.GatewayConfig.AdminSecret](#tykapisettingsgatewayconfigadminsecret)
            * [TykAPISettings.DashboardConfig.Endpoint](#tykapisettingsdashboardconfigendpoint)
            * [TykAPISettings.DashboardConfig.Port](#tykapisettingsdashboardconfigport)
            * [TykAPISettings.DashboardConfig.AdminSecret](#tykapisettingsdashboardconfigadminsecret)
         * [The profiles.json file](#the-profilesjson-file)
      * [Using Identity Providers](#using-identity-providers)
         * [Social](#social)
            * [Authenticate a user for the portal using Google and a constraint:](#authenticate-a-user-for-the-portal-using-google-and-a-constraint)
               * [What did we just do?](#what-did-we-just-do)
            * [Authenticate a user for the dashboard using Google and a constraint:](#authenticate-a-user-for-the-dashboard-using-google-and-a-constraint)
            * [OpenID Connect](#openid-connect)
            * [Salesforce](#salesforce)
            * [Create an OAuth token (with redirect) for users logging into your webapp or iOS app via Google:](#create-an-oauth-token-with-redirect-for-users-logging-into-your-webapp-or-ios-app-via-google)
         * [LDAP](#ldap)
            * [Log into Tyk Dashboard using LDAP](#log-into-tyk-dashboard-using-ldap)
            * [Log into Tyk Portal using LDAP](#log-into-tyk-portal-using-ldap)
            * [Generate an OAuth Token using an LDAP login form](#generate-an-oauth-token-using-an-ldap-login-form)
         * [Proxy Identity Provider](#proxy-identity-provider)
            * [JSON Data and Usernames](#json-data-and-usernames)
            * [Logging into the dashboard using a proxy provider](#logging-into-the-dashboard-using-a-proxy-provider)
            * [Generating a standard Auth Token using a Proxy Provider](#generating-a-standard-auth-token-using-a-proxy-provider)
         * [SAML](#saml)
            * [Logging into the dashboard using SAML](#logging-into-the-dashboard-using-saml)
            * [Log into Tyk Portal using SAML](#logging-into-tyk-portal-using-saml)
            * [Generating a standard Auth Token using SAML](#generating-a-standard-auth-token-using-saml)
      * [The Broker API](#the-broker-api)
         * [List profiles](#list-profiles)
         * [Add profile](#add-profile)
            * [Request](#request)
            * [Response](#response)
         * [Update profile](#update-profile)
            * [Request](#request-1)
            * [Response](#response-1)
         * [Delete profile](#delete-profile)
            * [Request](#request-2)
            * [Response](#response-2)
         * [Save profiles to disk](#save-profiles-to-disk)
            * [Request](#request-3)
            * [Response](#response-3)
            * [Outcome:](#outcome)

Tyk Identity Broker (TIB)
==============================

## What is the Tyk Identity Broker?

The Tyk Identity Broker provides a service-level component that enables delegated identities to be authorized and provide authenticated access to various Tyk-powered components such as the Tyk Dashboard, the Tyk Developer Portal and Tyk Gateway API flows such as OAuth access tokens, and regular API tokens.

![image](https://user-images.githubusercontent.com/14009/109294803-bbdc1400-783e-11eb-8d5c-640a2d944399.png)


### Requirements and dependencies

TIB requires:

- Tyk Gateway v1.9.1+
- Redis
- Tyk Dashboard v0.9.7.1+ (Only if you want to do SSO to Tyk Dashbaord UI or Tyk Developer Portal)

### Installation

You can install via Docker https://hub.docker.com/r/tykio/tyk-identity-broker/

Of via packages (deb or rpm): https://packagecloud.io/tyk/tyk-identity-broker/install#bash-deb

#### Run via Docker
To run the container, you can use the following command (assuming that you run it from the directory which contains `tib.conf` and `profiles.json` files, which **must** be edited according to [How to configure TIB](#how-to-configure-tib) section before running it):

```
docker run -p 3010:3010 -v $(pwd)/tib.conf:/opt/tyk-identity-broker/tib.conf -v $(pwd)/profiles.json:/opt/tyk-identity-broker/profiles.json tykio/tyk-identity-broker
```

### Usage

No command line arguments are needed, but if you are running TIB from another dir or during startup, you will need to set the absolute paths to the profile and config files

	Usage of ./tyk-auth-proxy:
	  -c=string
			Path to the config file (default "tib.conf")
	  -p#=string
			Path to the profiles file (default "profiles.json")

### Log level
You set the log level using the environment variable `TYK_LOGLEVEL`

Possible levels: `"debug"` , `"error"`, `"warn"` and `"info"` which is also the default.

For instance for debug `export TYK_LOGLEVEL=debug`

### How it works

Tyk Identity Broker provides a simple API, which traffic can be sent *through*, the API will match the request to a *profile* which then exposes two things:

- An Identity Provider - the thing that will authorize a user and validate their identity
- An Identity Handler - the thing that will authenticate a user with a delegated service (in this case, Tyk)

#### Identity Providers

Identity providers can be anything, so long as they implement the `tap.TAProvider` interface. Bundled with TIB at the moment you have three providers:

1. Social - Provides OAuth handlers for many popular social logins (such as Google, Github and Bitbucket), as well as general OpenID Connect support
2. LDAP - A simple LDAP protocol binder that can validate a username and password against an LDAP server (tested against OpenLDAP)
3. Proxy - A generic proxy handler that will forward a request to a third party and provides multiple "validators" to identify whether a response is successful or not (e.g. status code, content match and regex)
4. SAML - Provides a way to authenticate against a SAML IDP.

#### Identity Handlers

An identity handler will perform a predefined set of actions once a provider has validated an identity, these actions are defined as a set of action types:

**Pass through or redirect user-based actions**

- `GenerateOrLoginDeveloperProfile` - Will create or login a user to the Tyk Developer Portal
- `GenerateOrLoginUserProfile`  - Will log a user into the dashboard (this does not create a user, only drops a temporary session for the user to have access)
- `GenerateOAuthTokenForClient` - Will act as a client ID delegate and grant an Tyk-provided OAuth token for a user using a fragment in the redirect URL (standard flow)

** Direct or redirect **
- `GenerateTemporaryAuthToken` - Will generate a Tyk standard access token for the user, can be delivered as a redirect fragment OR as a direct API response (JSON)

These are actions are all handled by the `tap.providers.TykIdentityHandler` module which wraps the Tyk Gateway, Dashboard and Admin APIs to grant access to a stack.

Handlers are not limited to Tyk, a handler can be added quite easily by implementing the `TAProvider` so long as it implements this pattern and is registered it can handle any of the above actions for it's own target.

## How to configure TIB

Tyk Identity Broker is configured through two files: The configuration file (tib.conf) and the profiles file (profiles.json). TIB can also be managed via the REST API (detailed below) for automated configurations.

### The `tib.conf` file

```
{
	"Secret": "test-secret",
	"HttpServerOptions": {
		"UseSSL": true,
		"CertFile": "./certs/server.pem",
		"KeyFile": "./certs/server.key"
	},
	"SSLInsecureSkipVerify": true,
	"BackEnd": {
		"Name": "in_memory",
		"IdentityBackendSettings": {
			"Hosts" : {
				"localhost": "6379"
			},
			"Password": "",
			"Database": 0,
			"EnableCluster": false,
			"MaxIdle": 1000,
			"MaxActive": 2000
		}
	},
	"TykAPISettings": {
		"GatewayConfig": {
			"Endpoint": "http://{GATEWAY-DOMAIN}",
			"Port": "80",
			"AdminSecret": "{GATEWAY-SECRET}"
		},
		"DashboardConfig": {
			"Endpoint": "http://{DASHBOARD-DOMAIN}",
			"Port": "3000",
			"AdminSecret": "{ADMIN-DASHBOARD-SECRET}"
		}
	}
}
```

The various configuration options are outlined below:

#### `Secret`

The Gateway API secret to configure the Tyk Identity Broker remotely.

#### `HttpServerOptions.UseSSL`

Set this to `true` to turn on SSL for the server, this is *highly recommended*.

#### `HttpServerOptions.CertFile`

The path to the certificate file for this server, required for SSL

#### `HttpServerOptions.KeyFile`

The path to the key file for this server, required for SSL

#### `SSLInsecureSkipVerify`

If you run a local IDP, like Ping, with an untrusted SSL certificate, you can now turn off the client SSL verification by setting `SSLInsecureSkipVerify` to `true`.
This is useful when using OpenID Connect (OIDC). During the authorization there are calls to the `https://{IDP-DOMAIN}/.well-know/openid-configuration` and other endpoints to avoid error in case the certificate was signed by unknown authority

#### `BackEnd`

TIB is quite modular and different back-ends can be generated quite easily, out of the Box, TIB will store profile configurations in memory, which does not require any new configuration.

For Identity Handlers that provide token-based access, it is possible to enforce a "One token per provider, per user" policy, which keeps a cache of tokens assigned to identities in Redis, this is so that the broker can be scaled and share the cache across instances.

Since profiles are unlikely to change often, profiles are kept in-memory, but can be added, removed and modified using an API for automated setups if required.

#### `BackEnd.Hosts`

Add your Redis hosts here as a map of `hostname:port`. Since TIB uses the same cluster driver as Tyk, it is possible to have TIB interact with your existing Redis cluster if you enable it.

#### `BackEnd.Password`

The password for your Redis DB (recommended)

#### `BackEnd.Database`

If you are using multiple databases (not supported in Redis cluster), let TIB know which DB to use for Identity caching

#### `BackEnd.EnableCluster`

If you are using Redis cluster, enable it here to enable the slots mode

#### `BackEnd.MaxIdle`

Max idle connections to Redis

#### `BackEnd.MaxActive`

Max active Redis connections

#### `TykAPISettings`

This section enables you to configure the API credentials for the various Tyk Components TIB is interacting with.

#### `TykAPISettings.GatewayConfig.Endpoint`

The Hostname of the Tyk Gateway (this is for token generation purposes)

#### `TykAPISettings.GatewayConfig.Port`

The Port to use on the Tyk Gateway host

#### `TykAPISettings.GatewayConfig.AdminSecret`

The API secret for the Tyk Gateway API

#### `TykAPISettings.DashboardConfig.Endpoint`

The hostname of your Dashboard API

#### `TykAPISettings.DashboardConfig.Port`

The port of your Dashboard API

#### `TykAPISettings.DashboardConfig.AdminSecret`

The high-level secret for the Dashboard API. This is required because of the SSO-nature of some of the actions provided by TIB, it requires the capability to access a special SSO endpoint in the Dashboard Admin API to create one-time tokens for access.

### The `profiles.json` file

The Profiles configuration file outlines which identity providers to match to which handlers and what actions to perform. The entries in this file encapsulate the activity for a single endpoint based on the ID and provider name.

The file is JSON object which is essentially a list of objects:

```
[{
	"ActionType": "GenerateOrLoginUserProfile",
	"ID": "1",
	"IdentityHandlerConfig": {},
	"OrgID": "53ac07777cbb8c2d53000002",
	"CustomUserIDField": "FIELD-NAME",
	"ProviderConfig": {
		"CallbackBaseURL": "http://tib.domain.com:3010",
		"FailureRedirect": "http://tib.domain.com:3000/?fail=true",
		"UseProviders": [{
			"Key": "GOOGLE-OAUTH-TOKEN",
			"Name": "gplus",
			"Secret": "GOOGLE OAUTH SECRET",
			"SkipUserInfoRequest": false
		}]
	},
	"ProviderConstraints": {
		"Domain": "tyk.io",
		"Group": ""
	},
	"ProviderName": "SocialProvider",
	"ReturnURL": "http://tyk-dashboard.domain.com:3000/tap",
	"Type": "redirect"
}, {
	"ActionType": "GenerateOAuthTokenForClient",
	"ID": "2",
	"IdentityHandlerConfig": {
		"DashboardCredential": "ADVANCED-API-USER-API-TOKEN",
		"DisableOneTokenPerAPI": false,
		"OAuth": {
			"APIListenPath": "oauth-1",
			"BaseAPIID": "API-To-GRANT-ACCESS-TO",
			"ClientId": "TYK-OAUTH-CLIENT-ID",
			"RedirectURI": "http://your-app-domain.com/target-for-fragment",
			"ResponseType": "token",
			"Secret": "TYK-OAUTH-CLIENT-SECRET"
		}
	},
	"MatchedPolicyID": "POLICY-TO-ATTACH-TO-KEY",
	"OrgID": "53ac07777cbb8c2d53000002",
	"ProviderConfig": {
		"FailureRedirect": "http://yourdomain.com/failure-url",
		"LDAPAttributes": [],
		"LDAPUseSsl": false,
		"LDAPBaseDN": "cn=dashboard,ou=Group,dc=ldap,dc=tyk-test,dc=com",
		"LDAPEmailAttribute": "mail",
		"LDAPSearchScope": 2,
		"LDAPFilter": "((objectCategory=person)(objectClass=user)(cn=*USERNAME*))",
		"LDAPPort": "389",
		"LDAPServer": "localhost",
		"LDAPUserDN": "cn=*USERNAME*,cn=dashboard,ou=Group,dc=ldap,dc=tyk-test,dc=com"
	},
	"ProviderName": "ADProvider",
	"ReturnURL": "",
	"Type": "passthrough"
}]
```

Each item in a policy list dictates how that component will behave with the underlying services it is trying to talk to. In the above two examples, we have a social provider, that will allow Dashboard access to Google plus users that are part of the "tyk.io" domain. In the second example, we are generating an OAuth token for users that are validated via an LDAP server.

`DashboardCredential` - The credential of the dashboard user (that are used to login the UI or in the Dashboard API endpoints, not for the Admin Dashboard APIs)

In the following sections we outline multiple configurations you can use for Identity Provision and Handling

## Using Identity Providers

Tyk Identity Broker comes with a few providers built-in, these are specialized around a few use cases, but we focus primarily on:

- Enabling easy access via social logins to the developer portal (e.g. GitHub login)
- Enabling internal access to the dashboard (e.g. via LDAP/ActiveDirectory)
- Enabling easy token generation from a third party for things such as mobile apps and webapps without complex configuration

This next section outlines how to configure the various built-in providers.

### Social

The social provider is a thin wrapper around the excellent `goth` social auth library, modified slightly to work with a multi-tenant structure. The social provider should provide seamless integration with:

- Bitbucket
- Digital Ocean
- Dropbox
- Facebook
- GitHub
- Google+
- Linkedin
- Twitter
- SalesForce
- Any OpenID Connect provider

The social provider is ideal for SSO-style logins for the dashboard or for the portal, for certain providers (mainly Google+), where email addresses are returned as part for the user data, a constraint can be added to validate the users domain. This is useful for Google For Business Apss users that want to grant access to their domain users for the dashboard.

We've outlined a series of example configurations below for use with the social handler.

#### Authenticate a user for the portal using Google and a constraint:

The first thing to do with any social provider implementation is to make sure the OAuth client has been set up with the provider, and that the OAuth client has been set up with the correct callback URI.

**Step 1** Set up an OAuth client with google apps

1. Go to the [Google Developer Console](https://console.developers.google.com/) and create a new app
2. Register a new OAuth client, lets call it WebApp 1 (Select "New Credentials -> OAuth Client ID")
3. Select Web App
4. Add the following URL (modified for your domain) to the "Authorized redirect URIs" section: `http://tib-hostname:TIB-PORT/auth/{PROFILE-ID}/gplus/callback`

Save the client and take note of the secret and ID.

##### What did we just do?

We created a new OAuth client in Google apps that has a registered call back URL for TIB, the callback is very important, as this is how Google will tell TIB about the user logging in, the callback URI is constructed as follows:

	http://{TIB-HOST}:{TIB-PORT}/auth/{PROFILE-ID}/{PROVIDER-CODE}/callback

If you were to use twitter with a profile ID of 15, you would have a callback for twitter that looks like this:

	http://{TIB-HOST}:{TIB-PORT}/auth/15/twitter/callback

**Step 2** Create a profile object in profiles.json:

```
[{
	"ActionType": "GenerateOrLoginDeveloperProfile",
	"ID": "1",
	"IdentityHandlerConfig": {
		"DashboardCredential": "YOUR-DASHBOARD-USER-API-KEY"
	},
	"OrgID": "YOUR-ORG-ID",
	"ProviderConfig": {
		"CallbackBaseURL": "http://{TIB-HOST}:{TIB-PORT}",
		"FailureRedirect": "http://{PORTAL-DOMAIN}:{PORTAL-PORT}/portal/login/",
		"UseProviders": [{
			"Name": "gplus",
			"Key": "GOOGLE-OAUTH-CLIENT-KEY",
			"Secret": "GOOGLE-OAUTH-CLIENT-SECRET"
		}]
	},
	"ProviderConstraints": {
		"Domain": "yourdomain.com",
		"Group": ""
	},
	"ProviderName": "SocialProvider",
	"ReturnURL": "http://{PORTAL-DOMAIN}:{PORTAL-PORT}/portal/sso/",
	"Type": "redirect"
}]
```

This profile basically tells TIB to load a profile into memory with the ID of 1, that it should login or generate a developer profile via Google Plus, and it should only allow users from `yourdomain.com` domain-based email accounts.

The Return URL here is important, and is only provided in the *latest* version of Tyk Dashboard, as it makes use of new API endpoints to generate the SSO tokens required to allow remote access.

If your portal is configured under a different root (e.g. `/`, then replace the `/portal' component of the URLs with that of your actual portal.)

**Step 3 - Make a request to your TIB endpoint in your browser**

Now, start TIB by entering:

	./tib

And then point your browser at:

	http://{TIB-HOST}:{TIB-PORT}/auth/1/gplus

You will be asked to log into your account (make sure it is one that satisfies the constraints!), and once logged in, you should be redirected back via the TIB proxy to your portal, as a logged in user.

This user will be created with some user profile data, the user can edit and change their email address, but continue to log in with the same Google account (this data is stored separately).

#### Authenticate a user for the dashboard using Google and a constraint:

Similarly to the above, if we have our callback URL and client IDs set up with Google, we can use the following profile setup to access our developer portal using a social provider:

```
{
	"ActionType": "GenerateOrLoginUserProfile",
	"ID": "2",
	"IdentityHandlerConfig": null,
	"MatchedPolicyID": "1C",
	"OrgID": "53ac07777cbb8c2d53000002",
	"ProviderConfig": {
		"CallbackBaseURL": "http://\:{TIB-PORT}",
		"FailureRedirect": "http://{DASH-DOMAIN}:{DASH-PORT}/?fail=true",
		"UseProviders": [{
			"Name": "gplus",
			"Key": "GOOGLE-OAUTH-CLIENT-KEY",
			"Secret": "GOOGLE-OAUTH-CLIENT-SECRET"
		}]
	},
	"ProviderConstraints": {
		"Domain": "yourdomain.com",
		"Group": ""
	},
	"ProviderName": "SocialProvider",
	"ReturnURL": "http://{DASH-DOMAIN}:{DASH-PORT}/tap",
	"Type": "redirect"
}
```

It is worth noting in the above configuration that the return URL's have changed for failure and return states.

The login to the portal, much like the login to the dashboard, makes use of a one-time nonce to log the user in to the session. The nonce is only accessible for a few seconds. It is recommended that in production use, all of these transactions happen over secure SSL connections to avoid MITM snooping.

#### OpenID Connect
Similar to Google or Twitter auth, you can configure TIB to work with any OpenID Connect provider, like Okta, Ping Federate, or anything else. Just in addition to Key and Secret you need to provide Discovery URL, which you should find in documentation of your OpenID provider. Below is example configuration of Okta integration:

```
	"ProviderConfig": {
		"CallbackBaseURL": "http://{TIB-HOST}:{TIB-PORT}",
		"FailureRedirect": "http://{PORTAL-DOMAIN}:{PORTAL-PORT}/portal/login/",
		"UseProviders": [{
			"Name": "openid-connect",
			"Key": "OKTA-CLIENT-KEY",
			"Secret": "OKTA-CLIENT-SECRET",
			"Scopes": ["openid", "email"],
			"DiscoverURL": "https://<your-okta-domain>/.well-known/openid-configuration",
			"SkipUserInfoRequest": false
		}]
	},
```

By default, TIB uses the value of the subject field, returned by UserInfo Endpoint, to generate userId for Dashboard SSO.
Please ensure that your Identity Provider is not returning `URL` in subject field. If that's the case you should specify another field which uniquely identifies the user. You can specify the field name by setting `CustomUserIDField` in profile.json file.

TIB can also be configured to use the `openid email` claim.  This claim must be requested for Portal SSO to work, but it can also be used for Dashboard SSO.  Some OpenID providers return this claim by default, but not all, in which case the `openid email` claim needs to be included, as per the example above.

If you are getting 403 error, it can be that your OpenID provider requires providing client_id and secret_id via token url instead of basic http auth, and you need to add `"DisableAuthHeader": true` option to your provider configuration in "UseProviders" section.

Some Identity providers do not have support for `userinfo` endpoint, so you can optionally disable it using `SkipUserInfoRequest` flag, and rely only on information inside the ID token.

#### Salesforce

Similar to other social accounts, you can add support of Salesforce by specifying Provider Name as `salesforce`.

```
{
	"ProviderConfig": {
		"CallbackBaseURL": "http://tib.domain.com:3010",
		"FailureRedirect": "http://tib.domain.com:3000/?fail=true",
		"UseProviders": [{
			"Name": "salesforce",
			"Key": "SF-CLIENT-KEY",
			"Secret": "SF-CLIENT-SECRET",
		}]
	},
}
```

SSO for Salesforce Community is handled differently. Instead of using `salesforce` provider type, you should use `openid-connect` and set `user_id` in `CustomUserIDField` field. Here is sample profile.json file for community

```
{
	"CustomUserIDField": "user_id",
	"ProviderConfig": {
		"CallbackBaseURL": "http://tib.domain.com:3010",
		"FailureRedirect": "http://tib.domain.com:3000/?fail=true",
		"UseProviders": [{
			"Name": "openid-connect",
			"Key": "SF-CLIENT-KEY",
			"Secret": "SF-CLIENT-SECRET",
			"DiscoverURL": "https://community_url/.well-known/openid-configuration"
		}]
	},
}
```

#### Create an OAuth token (with redirect) for users logging into your webapp or iOS app via Google:

A common use case for Tyk Gateway users is to enable users to log into a web app or mobile app using a social provider such as Google, but have that user use a token in the app that is time-delimited and issued by their own API (or in this case, Tyk).

Tyk can act as an OAuth provider, but requires some glue code to work, in particular, generating a token based on the authentication of a third party, which needs to run on a server hosted by the owner of the application. This is not ideal in many scenarios where authentication has been delegated to a third-party provider (such as Google or GitHub).

In this case, we can enable this flow with Tyk Gateway by Using TIB.

What the broker will do is essentially the final leg of the authentication process without any new code, simply sending the user via TIB to the provider will suffice for them to be granted an OAuth token once they have authenticated in a standard, expected OAuth pattern.

Assuming we hav created an client ID and secret in Google Apps to grant us access to the users data, we need those details, and some additional ones from Tyk itself:

**Step 1 - Create an OAuth Client in Tyk Dashboard**

TIB will use the OAuth credentials for GPlus to access and authenticate the user, it will then use another set of client credentials to make the request to Tyk to generate a token response and redirect the user, this means we need to create an OAuth client in Tyk Dashboard before we can proceed.

One quirk with the Tyk API is that requests for tokens go via the base APIs listen path (`{listen_path}/oauth/authorize`), so we will need to know the listen path and ID of this API so TIB can make the correct API calls on your behalf.


```
{
	"ActionType": "GenerateOAuthTokenForClient",
	"ID": "3",
	"IdentityHandlerConfig": {
		"DashboardCredential": "{DASHBOARD-API-ID}",
		"DisableOneTokenPerAPI": false,
		"OAuth": {
			"APIListenPath": "{API-LISTEN-PATH}",
			"BaseAPIID": "{BASE-API-ID}",
			"ClientId": "{TYK-OAUTH-CLIENT-ID}",
			"RedirectURI": "http://{APP-DOMAIN}:{PORT}/{AUTH-SUCCESS-PATH}",
			"ResponseType": "token",
			"Secret": "{TYK-OAUTH-CLIENT-SECRET}"
		}
	},
	"MatchedPolicyID": "567a86f630c55e3256000003",
	"OrgID": "53ac07777cbb8c2d53000002",
	"ProviderConfig": {
		"CallbackBaseURL": "http://{TIB-DOMAIN}:{TIB-PORT}",
		"FailureRedirect": "http://{PORTAL-DOMAIN}:{PORTAL-PORT}/portal/login/?fail=true",
		"UseProviders": [{
			"Key": "GOOGLE-OAUTH-CLIENT-KEY",
			"Name": "gplus",
			"Secret": "GOOGLE-OAUTH-CLIENT-SECRET"
		}]
	},
	"ProviderName": "SocialProvider",
	"ReturnURL": "",
	"Type": "redirect"
}
```

There's a few new things here we need to take into account:

- `API-LISTEN-PATH` - This is the listen path of your API, TIB uses this to generate the OAuth token
- `BASE-API-ID` - The base API ID for the listen path mentioned earlier, this forms the basic access grant for the token (this will be superseded by the `MatchedPolicyID`, but is required for token generation)
- `TYK-OAUTH-CLIENT-ID` - The client ID for this profile within Tyk Gateway
- `TYK-OAUTH-CLIENT-SECRET` - The Client secret for this profile in Tyk Gateway
- `RedirectURI: http://{APP-DOMAIN}:{PORT}/{AUTH-SUCCESS-PATH}` - The Redirect URL set for this profile in the Tyk Gateway
- `ResponseType` - This can be `token` or `authorization_code`, the first will generate a token directly, the second will generate an auth code for follow up access. For SPWA and Mobile Apps it is recommended to just use `token`

When TIB successfully authorizes the user, and generates the token using the relevant OAuth credentials, it will redirect the user to the relevant redirect with their token or auth code as a fragment in the URl for the app to decode and use as needed.

There is a simplified flow which does not require a corresponding OAuth client in Tyk Gateway, and can just generate a standard token with the same flow.

### LDAP

The LDAP Identity Provider is experimental currently and provides limited functionality to bind a user to an LDAP server based on a username and password configuration. The LDAP provider currently does not extract user data from the server to populate a user object, but will provide enough defaults to work with all handlers.

#### Log into Tyk Dashboard using LDAP

```
{
	"ActionType": "GenerateOrLoginUserProfile",
	"ID": "4",
	"OrgID": "{YOUR-ORG-ID}",
	"ProviderConfig": {
		"LDAPUseSSL": false,
		"FailureRedirect": "http://{DASH-DOMAIN}:{DASH-PORT}/?fail=true",
		"LDAPAttributes": [],
		"LDAPPort": "389",
		"LDAPServer": "localhost",
		"LDAPUserDN": "cn=*USERNAME*,cn=dashboard,ou=Group,dc=test-ldap,dc=tyk,dc=io"
	},
	"ProviderName": "ADProvider",
	"ReturnURL": "http://{DASH-DOMAIN}:{DASH-PORT}/tap",
	"Type": "passthrough"
}
```

TIB can pull a username and password out of a request in two ways:

1. Two form fields called "username" and "password"
2. A basic auth header using the Basic Authentication standard form

By default, TIB will look for the two form fields. To enable Basic Auth header extraction, add `"GetAuthFromBAHeader": true` to the `ProviderConfig` section.

The request should be a `POST`. Example curl command can look like: `curl -X POST localhost:3010/auth/4/callback -F username=bob -F password=secret`

Set `LDAPUseSSL` to `true` if you want to use LDAPS (LDAP over SSL).

If you make this request with a valid user that can bind to the LDAP server, Tyk will redirect the user to the dashboard with a valid session. There's no more to it, this mechanism is pass-through and is transparent to the user, with TIB acting as a direct client to the LDAP provider.

**Note** The `LDAPUserDN` field MUST contain the special `*USERNAME*` marker in order to construct the users OU properly.

#### Log into Tyk Portal using LDAP

LDAP requires little configuration, we can use the same provider config above, with one that logs us into the portal instead - notice the change in the handler configuration and the return URL:

```
{
	"ActionType": "GenerateOrLoginDeveloperProfile",
	"ID": "5",
	"IdentityHandlerConfig": {
		"DashboardCredential": "822f2b1c75dc4a4a522944caa757976a"
	},
	"OrgID": "53ac07777cbb8c2d53000002",
	"ProviderConfig": {
		"FailureRedirect": "http://{PORTAL-DOMAIN}:{PORTAL-PORT}/portal/login/",
		"LDAPAttributes": [],
		"LDAPUseSSL": false,
		"LDAPPort": "389",
		"LDAPServer": "localhost",
		"LDAPUserDN": "cn=*USERNAME*,cn=dashboard,ou=Group,dc=test-ldap,dc=tyk,dc=io"
	},
	"ProviderConstraints": {
		"Domain": "",
		"Group": ""
	},
	"ProviderName": "ADProvider",
	"ReturnURL": "http://{PORTAL-DOMAIN}:{PORTAL-PORT}/portal/sso/",
	"Type": "passthrough"
}
```

#### Generate an OAuth Token using an LDAP login form

The configuration below will take a request that is posted to TIB, authenticate it against LDAP, if the request is valid, it will redirect to the Tyk Gateway clients Redirect URI with the token as a URL fragment:

```
{
	"ActionType": "GenerateOAuthTokenForClient",
	"ID": "6",
	"IdentityHandlerConfig": {
		"DashboardCredential": "{DASHBAORD-API-ID}",
		"DisableOneTokenPerAPI": false,
		"OAuth": {
			"APIListenPath": "{API-LISTEN-PATH}",
			"BaseAPIID": "{BASE-API-ID}",
			"ClientId": "{TYK-OAUTH-CLIENT-ID}",
			"RedirectURI": "http://{APP-DOMAIN}:{PORT}/{AUTH-SUCCESS-PATH}",
			"ResponseType": "token",
			"Secret": "{TYK-OAUTH-CLIENT-SECRET}"
		}
	},
	"MatchedPolicyID": "POLICY-ID",
	"OrgID": "53ac07777cbb8c2d53000002",
	"ProviderConfig": {
		"FailureRedirect": "http://{APP-DOMAIN}:{PORT}/failure",
		"LDAPAttributes": [],
		"LDAPUseSSL": false,
		"LDAPPort": "389",
		"LDAPServer": "localhost",
		"LDAPUserDN": "cn=*USERNAME*,cn=dashboard,ou=Group,dc=ldap,dc=tyk-ldap-test,dc=com"
	}
	"ProviderName": "ADProvider",
	"ReturnURL": "",
	"Type": "passthrough"
}
```

#### Using two phase LDAP authentication

In some cases only privileged users are allowed perform LDAP search. In this case you can specify your admin user using `LDAPAdminUser` and `LDAPAdminPassword` options. TIB will perform initial bind as admin user, then will perform a LDAP lookup based on specified DN template or `LDAPFilter`, and will do bind one more time, with user DN.

```
{
    "ActionType": "GenerateOrLoginUserProfile",
    "ID": "4",
    "OrgID": "59fc80d9158519599ca23cfc",
    "ProviderConfig": {
        "FailureRedirect": "https://tyk-dashboard:3000/?fail=true",
        "LDAPPort": "389",
        "LDAPAdminUser": "admin",
        "LDAPAdminPassword": "password",
        "LDAPServer": "localhost",
        "LDAPUserDN": "uid=*USERNAME*,dc=example,dc=org"
    },
    "ProviderName": "ADProvider",
    "ReturnURL": "https://tyk-dashboard:3000/tap",
    "Type": "passthrough"
}
```

### Proxy Identity Provider

The proxy identity provider is a generic solution to more legacy problems, as well as a way to handle flows such as basic auth access with third party providers or OAuth password grants where the request can just be passed through to the providing endpoint to return a direct response.

The proxy provider will take a request, proxy it to an upstream host, capture the response, and analyze it for triggers of "success", if the triggers come out as true, then the provider will treat the request as authenticated and hand over to the Identity Handler to perform whatever action is required with the user data.

Success can be triggered using three methods:

1. Response code - e.g. if this is an API request, a simple "200" response would suffice to act as a successful authentication
2. Response body exact match - You can have a base64 encoded body that you would expect as a successful match, if the two bodies are the same, then the request will be deemed successful
3. Regex - Most likely, the response might be dynamic (and return a response code, timestamp or other often changing parameter), in which case you may want to just match the response to a regex.

These can be used in conjunction as gates, e.g. a response must be 200 OK and match the regex in order to be marked as successful.

#### JSON Data and Usernames

The Proxy provider can do some clever things, such as extract JSON data from the response and decode it, as well as pull username data from the Basic Auth header (for example, if your identity provider supports dynamic basic auth).

#### Logging into the dashboard using a proxy provider

```
{
	"ActionType": "GenerateOrLoginUserProfile",
	"ID": "7",
	"OrgID": "{YOUR-ORG-ID}",
	"ProviderConfig": {
		"AccessTokenField": "access_token",
		"ExrtactUserNameFromBasicAuthHeader": false,
		"OKCode": 200,
		"OKRegex": "",
		"OKResponse": "",
		"ResponseIsJson": true,
		"TargetHost": "http://{TARGET-HOSTNAME}:{PORT}/",
		"UsernameField": "user_name"
	},
	"ProviderName": "ProxyProvider",
	"ReturnURL": "http://{DASH-DOMAIN}:{DASH-PORT}/tap",
	"Type": "redirect"
}
```

#### Generating a standard Auth Token using a Proxy Provider

```
{
	"ActionType": "GenerateTemporaryAuthToken",
	"ID": "8",
	"IdentityHandlerConfig": {
		"DashboardCredential": "822f2b1c75dc4a4a522944caa757976a",
		"DisableOneTokenPerAPI": false,
		"TokenAuth": {
			"BaseAPIID": "e1d21f942ec746ed416ab97fe1bf07e8"
		}
	},
	"MatchedPolicyID": "5654566b30c55e3904000003",
	"OrgID": "53ac07777cbb8c2d53000002",
	"ProviderConfig": {
		"AccessTokenField": "access_token",
		"ExrtactUserNameFromBasicAuthHeader": false,
		"OKCode": 200,
		"OKRegex": "",
		"OKResponse": "",
		"ResponseIsJson": true,
		"TargetHost": "http://{TARGET-HOSTNAME}:{PORT}/",
		"UsernameField": "user_name"
	},
	"ProviderName": "ProxyProvider",
	"ReturnURL": "",
	"Type": "passthrough"
}
```
### SAML
SAML authentication is a way for a service provider, such as the Tyk Dashboard or Portal, to assert the Identity of a User via a third party.

Tyk Identity Broker can act as the go-between for the Tyk Dashboard and Portal and a third party identity provider. Tyk Identity broker can also interpret and pass along information about the user who is logging in such as Name, Email and group or role metadata for enforcing role based access control in the Tyk Dashboard.

###### Gateway & Dashboard pre-requisites
The Gateway will encode the certificate being used for SAML at the time of upload. In order for the dashboard to be able to use this certificate you will need to match the `secret` within tyk.conf with the `tyk_api_config.secret` tyk_analytics.conf. Alternatively you can enable `security.private_certificate_encoding_secret` in both tyk.conf for the gateway and tyk_analytics.conf for the dashboard.

###### SAML Glossary
SAML metadata is an XML document which contains information necessary for interaction with SAML-enabled identity or service providers. The document contains e.g. URLs of endpoints, information about supported bindings, identifiers and public keys. Once you create your TIB profile you can find the SP metadata file under `{Dashboard HOST}/auth/{TIB Profile Name}/saml/metadata`  

The provider config for SAML has the following values that can be configured in a Profile:

`SAMLBaseURL`: The host of TIB that will be used in the metadata document for the Service Provider. This will form part of the metadata URL used as the Entity ID by the IDP. The redirects configured in the IDP must match the expected Host and URI configured in the metadata document made available by Tyk Identity Broker.

`FailureRedirect`: Where to redirect failed login requests.

`IDPMetaDataURL`: The metadata URL of your IDP which will provide Tyk Identity Broker with information about the IDP such as EntityID, Endpoints (Single Sign On Service Endpoint, Single Logout Service Endpoint), its public X.509 cert, NameId Format, Organization info and Contact info. This metadata XML can be signed providing a public X.509 cert and the private key.

`CertLocation`: An X.509 certificate and the private key for signing your requests to the IDP, this should be one single file with the cert and key concatenated.

`ForceAuthentication`: Ignore any session held by the IDP and force re-login every request.

`SAMLEmailClaim`: Key for looking up the email claim in the SAML assertion form the IDP. Defaults to: http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress

`SAMLForenameClaim`: Key for looking up the forename claim in the SAML assertion form the IDP. Defaults to: http://schemas.xmlsoap.org/ws/2005/05/identity/claims/forename

`SAMLSurnameClaim`: Key for looking up the surname claim in the SAML assertion form the IDP. Defaults to: http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname

Example profile configuration:

```json
{
    "ActionType": "GenerateOrLoginUserProfile",
    "ID": "saml-sso-login",
    "OrgID": "{YOUR_ORGANISATION_ID}",
    "CustomEmailField": "",
    "IdentityHandlerConfig": {
        "DashboardCredential": "{DASHBOARD_USER_API_KEY}"
    },
    "ProviderConfig": {
        "SAMLBaseURL": "https://{HOST}",
        "FailureRedirect": "http://{DASHBOARD_HOST}:{PORT}/?fail=true",
        "IDPMetaDataURL": "{IDP_METADATA_URL}",
        "CertLocation": "myservice.cert",
        "ForceAuthentication": false,
        "SAMLEmailClaim": "",
        "SAMLForenameClaim": "",
        "SAMLSurnameClaim": ""
    },
    "ProviderName": "SAMLProvider",
    "ReturnURL": "http://{DASHBOARD_URL}:{PORT}/tap",
    "Type": "redirect"
}
```

#### Logging into the dashboard using SAML

In order to have dashboard access using SAML we need to create a TIB profile within our Tyk dashboard. You can find this under Identity Management on the left control pane. You can then select create profile, and use the raw editor to copy the example below:

```json
{
    "ID": "saml-sso-dash-login",
    "OrgID": {ORG-ID},
    "ActionType": "GenerateOrLoginUserProfile",
    "Type": "redirect",
    "ProviderName": "SAMLProvider",
    "ProviderConfig" : {
        "CertLocation": {CERT-PATH-OR-ID},
        "SAMLBaseURL": {TIB-HOST},
        "ForceAuthentication": false,
        "FailureRedirect": "{DASH-HOST}/?fail=true",
        "IDPMetaDataURL": {METADATA-URL-PROVIDED-BY-IDP}
    },
    "IdentityHandlerConfig" : {
        "DashboardCredential" : "{DASH-CREDENTIAL}"
    },
    "ReturnURL" : "http://tyk-dashboard:3000/tap"
}
```

#### Logging into Tyk Portal using SAML

To obtain tyk portal access it's similar to the profile above, the minimum configuration to get this access is defined as the next profile:

```json
{
    "ID": "saml-sso-dev-portal-login",
    "ActionType": "GenerateOrLoginDeveloperProfile",
    "OrgID": {ORG-ID},
    "ProviderConfig": {
        "SAMLBaseURL": {TIB-HOST},
        "FailureRedirect": "{PORTAL-HOST}/portal/login/",
        "IDPMetaDataURL": {METADATA-URL-PROVIDED-BY-IDP},
        "CertLocation": {CERT-PATH-OR-ID},
        "ForceAuthentication": true
    },
    "IdentityHandlerConfig": {
            "DashboardCredential": "{DASH-CREDENTIAL}"
    },
    "ProviderName": "SAMLProvider",
    "ReturnURL": {PORTAL-HOST}/sso/},
    "Type": "redirect"
}
```

#### Generating a Standard Auth Token using SAML

```json
  {
        "ID": "saml-for-auth-api-token",
        "OrgID": {ORG-ID},
        "ActionType": "GenerateTemporaryAuthToken",
        "MatchedPolicyID": {POLICY-ID},
        "Type": "passthrough",
        "ProviderName": "SAMLProvider",
        "ProviderConfig": {
            "CertLocation": {CERT-PATH-OR-ID},
            "ForceAuthentication": false,
            "IDPMetaDataURL": {METADATA-URL-PROVIDED-BY-IDP},
            "SAMLBaseURL": {TIB-HOST}
        },
        "IdentityHandlerConfig": {
            "DashboardCredential": {DASH-CREDENTIAL},
            "TokenAuth":{
                "BaseAPIID": {API-ID}
            }
        }
    }
```


#### User Group ID Support

You can specify IDP User Groups within a TIB Profile. This can either be a static or dynamic setting. You will first need to create a matching group within the Tyk Dashboard to correspond to the group coming form the IDP. Use the id of the group created within Tyk to map it to the IDP group name. In the example below the `"admin"` and `"analytics"` are the group names being passed by the IDP and `<admin-group-id>` as well as `<analytics-group-id>` are the id from the groups you have already created in Tyk to correspond to the groups being passed.

If you wish to map users via email make sure you have `"sso_enable_user_lookup": true,` within your tyk_analytics.conf file.

```
{
  "DefaultUserGroupID": "<dashboard-user-group-id>",
  "CustomUserGroupField": "scope",
  "UserGroupMapping": {
    "admin": "<admin-group-id>",
    "analytics": "<analytics-group-id>"
  }
}
```
When doing SSO for a user, you need to think about the user's permissions once they are logged into the application. During the SSO flow of a user, TIB can request Tyk-Dashboard to login that user with certain user group permissions. In order to configure the user's permission you need to create a group in the Dashboard and use this group object ID as a value for these fields:
* For a static setting  set `DefaultUserGroupID` with a Dashboard group id. TIB will use it as the default user permissions when requesting a nonce from the dashboard. **Note:** If you don't set this field, the user will be logged in as an admin dashboard user.
* For a dynamic setting based on OAuth/OpenID scope, use `CustomUserGroupField` with  `UserGroupMapping` listing your User Groups names from the scopes to user group IDs in the dashboard, in the following format - `"<user-group-name>": "<user-group-id>"`

## The Broker API

Tyk Identity Broker has a simple API to allow policies to be created, updated, removed and listed for programmatic and automated access. TIB also has a "flush" feature that enables you to flush the current configuration to disk for use when the client starts again.

TIB does not store profiles in shared store, so if you have multiple TIB instances, they need to be configured individually (for now), since we don't expect TIB stores to change often, this is acceptable.

### List profiles

```
GET /api/profiles/
Authorization: test-secret

{
	"Status": "ok",
	"ID": "",
	"Data": [
		{
			"ActionType": "GenerateTemporaryAuthToken",
			"ID": "11",
			"IdentityHandlerConfig": {
				"DashboardCredential": "822f2b1c75dc4a4a522944caa757976a",
				"DisableOneTokenPerAPI": false,
				"TokenAuth": {
					"BaseAPIID": "e1d21f942ec746ed416ab97fe1bf07e8"
				}
			},
			"MatchedPolicyID": "5654566b30c55e3904000003",
			"OrgID": "53ac07777cbb8c2d53000002",
			"ProviderConfig": {
				"ExrtactUserNameFromBasicAuthHeader": true,
				"OKCode": 200,
				"OKRegex": "origin",
				"OKResponse": "ewogICJvcmlnaW4iOiAiNjIuMjMyLjExNC4yNTAsIDE3OC42Mi4xMS42MiwgMTc4LjYyLjExLjYyIgp9Cg==",
				"TargetHost": "http://sharrow.tyk.io/ba-1/"
			},
			"ProviderConstraints": {
				"Domain": "",
				"Group": ""
			},
			"ProviderName": "ProxyProvider",
			"ReturnURL": "",
			"Type": "passthrough"
		},
		{
			"ActionType": "GenerateOAuthTokenForClient",
			"ID": "6",
			"IdentityHandlerConfig": {
				"DashboardCredential": "{DASHBAORD-API-ID}",
				"DisableOneTokenPerAPI": false,
				"OAuth": {
					"APIListenPath": "{API-LISTEN-PATH}",
					"BaseAPIID": "{BASE-API-ID}",
					"ClientId": "{TYK-OAUTH-CLIENT-ID}",
					"RedirectURI": "http://{APP-DOMAIN}:{PORT}/{AUTH-SUCCESS-PATH}",
					"ResponseType": "token",
					"Secret": "{TYK-OAUTH-CLIENT-SECRET}"
				}
			},
			"MatchedPolicyID": "POLICY-ID",
			"OrgID": "53ac07777cbb8c2d53000002",
			"ProviderConfig": {
				"FailureRedirect": "http://{APP-DOMAIN}:{PORT}/failure",
				"LDAPAttributes": [],
				"LDAPUseSSL": false,
				"LDAPPort": "389",
				"LDAPServer": "localhost",
				"LDAPUserDN": "cn=*USERNAME*,cn=dashboard,ou=Group,dc=ldap,dc=tyk-ldap-test,dc=com"
			}
			"ProviderName": "ADProvider",
			"ReturnURL": "",
			"Type": "passthrough"
		}
	]
}
```

### Add profile

#### Request

```
POST /api/profiles/{id}
Authorization: test-secret

{
			"ActionType": "GenerateTemporaryAuthToken",
			"ID": "11",
			"IdentityHandlerConfig": {
				"DashboardCredential": "822f2b1c75dc4a4a522944caa757976a",
				"DisableOneTokenPerAPI": false,
				"TokenAuth": {
					"BaseAPIID": "e1d21f942ec746ed416ab97fe1bf07e8"
				}
			},
			"MatchedPolicyID": "5654566b30c55e3904000003",
			"OrgID": "53ac07777cbb8c2d53000002",
			"ProviderConfig": {
				"ExrtactUserNameFromBasicAuthHeader": true,
				"OKCode": 200,
				"OKRegex": "origin",
				"OKResponse": "ewogICJvcmlnaW4iOiAiNjIuMjMyLjExNC4yNTAsIDE3OC42Mi4xMS42MiwgMTc4LjYyLjExLjYyIgp9Cg==",
				"TargetHost": "http://sharrow.tyk.io/ba-1/"
			},
			"ProviderConstraints": {
				"Domain": "",
				"Group": ""
			},
			"ProviderName": "ProxyProvider",
			"ReturnURL": "",
			"Type": "passthrough"
}
```

#### Response

```
{
	"Status": "ok",
	"ID": "11",
	"Data": {
		"ID": "11",
		"OrgID": "53ac07777cbb8c2d53000002",
		"ActionType": "GenerateTemporaryAuthToken",
		"MatchedPolicyID": "5654566b30c55e3904000003",
		"Type": "passthrough",
		"ProviderName": "ProxyProvider",
		"ProviderConfig": {
			"ExrtactUserNameFromBasicAuthHeader": true,
			"OKCode": 200,
			"OKRegex": "origin",
			"OKResponse": "ewogICJvcmlnaW4iOiAiNjIuMjMyLjExNC4yNTAsIDE3OC42Mi4xMS42MiwgMTc4LjYyLjExLjYyIgp9Cg==",
			"TargetHost": "http://sharrow.tyk.io/ba-1/"
		},
		"IdentityHandlerConfig": {
			"DashboardCredential": "822f2b1c75dc4a4a522944caa757976a",
			"DisableOneTokenPerAPI": false,
			"TokenAuth": {
				"BaseAPIID": "e1d21f942ec746ed416ab97fe1bf07e8"
			}
		},
		"ProviderConstraints": {
			"Domain": "",
			"Group": ""
		},
		"ReturnURL": ""
	}
}
```

### Update profile

#### Request

```
PUT /api/profiles/{id}
Authorization: test-secret

{
			"ActionType": "GenerateTemporaryAuthToken",
			"ID": "11",
			"IdentityHandlerConfig": {
				"DashboardCredential": "822f2b1c75dc4a4a522944caa757976a",
				"DisableOneTokenPerAPI": false,
				"TokenAuth": {
					"BaseAPIID": "e1d21f942ec746ed416ab97fe1bf07e8"
				}
			},
			"MatchedPolicyID": "5654566b30c55e3904000003",
			"OrgID": "53ac07777cbb8c2d53000002",
			"ProviderConfig": {
				"ExrtactUserNameFromBasicAuthHeader": true,
				"OKCode": 200,
				"OKRegex": "origin",
				"OKResponse": "ewogICJvcmlnaW4iOiAiNjIuMjMyLjExNC4yNTAsIDE3OC42Mi4xMS42MiwgMTc4LjYyLjExLjYyIgp9Cg==",
				"TargetHost": "http://sharrow.tyk.io/ba-1/"
			},
			"ProviderConstraints": {
				"Domain": "",
				"Group": ""
			},
			"ProviderName": "ProxyProvider",
			"ReturnURL": "",
			"Type": "passthrough"
}
```

#### Response

```
{
	"Status": "ok",
	"ID": "11",
	"Data": {
		"ID": "11",
		"OrgID": "53ac07777cbb8c2d53000002",
		"ActionType": "GenerateTemporaryAuthToken",
		"MatchedPolicyID": "5654566b30c55e3904000003",
		"Type": "passthrough",
		"ProviderName": "ProxyProvider",
		"ProviderConfig": {
			"ExrtactUserNameFromBasicAuthHeader": true,
			"OKCode": 200,
			"OKRegex": "origin",
			"OKResponse": "ewogICJvcmlnaW4iOiAiNjIuMjMyLjExNC4yNTAsIDE3OC42Mi4xMS42MiwgMTc4LjYyLjExLjYyIgp9Cg==",
			"TargetHost": "http://sharrow.tyk.io/ba-1/"
		},
		"IdentityHandlerConfig": {
			"DashboardCredential": "822f2b1c75dc4a4a522944caa757976a",
			"DisableOneTokenPerAPI": false,
			"TokenAuth": {
				"BaseAPIID": "e1d21f942ec746ed416ab97fe1bf07e8"
			}
		},
		"ProviderConstraints": {
			"Domain": "",
			"Group": ""
		},
		"ReturnURL": ""
	}
}
```

### Delete profile

#### Request

```
Delete /api/profiles/{id}
Authorization: test-secret

[empty body]

```

#### Response

```
{
	"Status": "ok",
	"ID": "200",
	"Data": {}
}
```

### Save profiles to disk

#### Request

```
POST /Authorization: test-secret
[empty body]api/profiles/save
```

#### Response

```
{
	"Status": "ok",
	"ID": "",
	"Data": {}
}
```

#### Outcome:

The existing profiles.json file will be backed up to a new file, and a the current profiles data in memory will be flushed to disk as the new profiles.json file. Backups are time stamped (e.g. `profiles_backup_1452677499.json`).
