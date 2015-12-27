Tyk Identity Broker (TIB)
=========================

## What is it?

The Tyk Identity Broker provides a service-level component that can enable delegated identities for various Tyk components. 

This is best described in an example:

### Use case 1 - Portal + Social

As a developer in an organisation that uses Google services (for email, calendar, docs etc.) I want to be able to log into and register for the Tyk Developer Portal using my Google account.

### Use case 2 - Portal + LDAP

As a developer in an organisation that ususes ActiveDirectory or LDAP to log into the Tyk developer portal.

### Use case 3 - Admin + Social

As a developer in an organisation that uses Google services (for email, calendar, docs etc.) I want to be able to log into the Tyk Dashboard admin using my Google account.

### Use case 3 - Admin + LDAO

As a developer in an organisation that uses ActiveDirectory or LDAP, I want to be able to log into the Tyk Dashboard admin using an LDAP filter.

### Use case 4 - OAuth token + social

As a developer I want to issue OAuth Access tokens to my iOS app users but have those tokens generated and throttled / managed by Tyk, but user should be able to log in using Google Plus, Facebook or Twitter

### Use case 5 - OAuth token + LDAP

As a developer I want to issue OAuth Access tokens to my internal developers but have those tokens generated and throttled / managed by Tyk, the user should be able to log in using their LDAP account.

## The skinny - TIB is a middleman

You can use TIB to hook up any identity provider:

- Bitbucket
- Digital Ocean
- Dropbox
- Facebook
- GitHub
- Google+
- Lastfm
- Linkedin
- Spotify
- Twitch
- Twitter

Or enterprise provider:

- LDAP
- SAML assertion (tbc)
- JSON Web Token claim validation (tbc)

And hook up an appropriate action to perform with Tyk Gateway that allows some kind of profile-specific access:

- Tyk Developer Portal Account Login
- Tyk Dashboard Login
- Generate OAuth Access Token based on a Tyk API Gateway Policy
- Generate OAuth Authorization code based on a Tyk API Gateway Policy (tbc)
- Generate a generec Access Token based on a Tyk API Gateway Policy (tbc)

## How do I use it?

(TODO)

## How can I contribute?

(TODO)