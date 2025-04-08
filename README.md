# CHECK24 GenDev Internet Provider Comparison Challenge

**Note:** This is the challenge for the **7th round** of the [GenDev Scholarship](https://check24.de/gen-dev). We look forward to your application - happy coding! 🤓

## Table of Contents

- [CHECK24 GenDev Internet Provider Comparison Challenge](#check24-gendev-internet-provider-comparison-challenge)
  - [Table of Contents](#table-of-contents)
  - [The Challenge 🤔](#the-challenge-)
    - [Minimum Requirements](#minimum-requirements)
    - [Optional Features](#optional-features)
    - [Getting Started](#getting-started)
  - [Providers 🛜](#providers-)
    - [WebWunder](#webwunder)
    - [ByteMe](#byteme)
    - [Ping Perfect](#ping-perfect)
    - [VerbynDich](#verbyndich)
    - [Servus Speed](#servus-speed)
  - [UI 💅](#ui-)
  - [What We're Looking For 🧐](#what-were-looking-for-)
  - [Submitting Your Project 🚀](#submitting-your-project-)
  - [Questions? ❓](#questions-)

## The Challenge 🤔

Build an application that allows users to compare internet providers. 🌐 You will receive five different API endpoints from five different internet providers, but these APIs may be unreliable.
Your goal is to ensure a smooth comparison experience despite potential API failures or slow responses, without negatively impacting the user experience or
limiting product variety while still displaying only actual offers that the internet providers are able to conclude.

### Minimum Requirements

1. Robust handling of API failures or delays.
2. Make sure to provide useful sorting and filtering options.
3. Your application should include a share link feature that allows users to share the results page with others via platforms like WhatsApp. Even if a provider is down, the share links should still direct users to the shared offers.
4. Credentials of the provider APIs should NEVER be leaked to the user.

### Optional Features

- Address autocompletion
- Validation of customer Input
- Session state to remember the user's last search

### Getting Started

You can register for the API access [here](https://register.gendev7.check24.fun/).

## Providers 🛜

### WebWunder

WebWunder provides a SOAP web service for internet offers. The service follows traditional SOAP protocols and requires XML-based request/response handling.

**API Details**

- **WSDL URL**: <https://webwunder.gendev7.check24.fun/endpunkte/soap/ws/getInternetOffers.wsdl>
- **Request Model**: `legacyGetInternetOffers` Element
- **Response Model**: `output` XML element
- To authenticate against the API, an `X-Api-Key` HTTP Header is needed

**Example Request**

```xml
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:gs="http://webwunder.gendev7.check24.fun/offerservice">
   <soapenv:Header/>
   <soapenv:Body>
      <gs:legacyGetInternetOffers>
         <gs:input>
            <gs:installation>true</gs:installation>
            <gs:connectionEnum>DSL</gs:connectionEnum>
            <gs:address>
               <gs:street>Example Street</gs:street>
               <gs:houseNumber>123</gs:houseNumber>
               <gs:city>Berlin</gs:city>
               <gs:plz>10115</gs:plz>
               <gs:countryCode>DE</gs:countryCode>
            </gs:address>
         </gs:input>
      </gs:legacyGetInternetOffers>
   </soapenv:Body>
</soapenv:Envelope>

```

### ByteMe

This provider outputs its offers in this CSV format:
`productId,providerName,speed,monthlyCostInCent,afterTwoYearsMonthlyCost,durationInMonths,connectionType,installationService,tv,limitFrom,maxAge,voucherType,voucherValue`

They recently informed about some issue with their API, where the same offers are sent out multiple times.

Authentication is done via the `X-Api-Key` header.

Their API can be reached under <https://byteme.gendev7.check24.fun/app/api/products/data>

The API endpoint takes the following string query params:

- street
- houseNumber
- city
- plz

### Ping Perfect

Ping Perfect's OpenAPI spec can be found [here](https://register.gendev7.check24.fun/openapi-pingperfect.json)
The Request body must be signed using a super secure custom signing method.

**Steps for Signature Calculation**

1. **Generate a Timestamp**
   - The current Unix epoch timestamp (in seconds) is used to ensure freshness.

2. **Prepare the Data to Sign**
   - The request body (JSON payload) is converted to a string.
   - Get the current time as UNIX timestamp in seconds.
   - Concatenate the timestamp and the request body string with `:` as a separator.

3. **Compute the HMAC-SHA256 Signature**
   - Use the `HMAC-SHA256` algorithm.
   - The `signatureSecret` (a shared secret key) is used as the HMAC key.
   - The resulting hash is converted to a hexadecimal string.

4. **Attach the Signature to the Request Headers**

- The calculated signature is added as `X-Signature` in the HTTP request headers.
- The timestamp is also included as `X-Timestamp` for verification.
- The `X-Client-Id` identifies the client making the request.

### VerbynDich

Sadly VerbynDich doesn't provide a proper API.
It's reachable at `https://verbyndich.gendev7.check24.fun/`.
The endpoint is at `/check24/data`.
The authentication is done via the `apiKey` query parameter.
The request body is a simple string with the address in the following format:
`street;house number;city;plz` with no newlines or whitespaces.
Also an optional parameter with the `page` is given. The page is an integer starting from 0.
The response is a JSON response with the following format:

```json
{
    "product": "string",
    "description": "string",
    "last": "boolean whether this is the last offer",
    "valid": "boolean whether it's a valid offer"
}
```

You will need to extract the necessary information from the description field.
Try multiple different addresses to get a good overview of the available offers and descriptions.

### Servus Speed

Servus Speed provides a RESTful API for internet offers. The auth is done via basic auth.
The OpenAPI spec can be found [here](https://register.gendev7.check24.fun/openapi-servus-speed.json).
Use the `/api/external/available-products` endpoint to get the available products.
After that you can use the `/api/external/product-details/{productId}` endpoint to get the details of a specific product.
Currently only the country code `DE` is supported. The discount is a fixed discount in Cent.

## UI 💅

The application should have a graphical user interface. It can be a web, mobile, or desktop application. The technology is up to you, but it should be simple and intuitive to use.
You can check out our [Internet Provider Comparison](https://www.check24.de/internet/) for some inspiration on what the UI could look like.

## What We're Looking For 🧐

- **Code Quality:** Clean, maintainable, and well-documented code
- **Performance:** Fast and responsive application, efficient algorithms
- **User Experience:** Fast and seamless comparison results regardless of API failures
- **Creativity:** Additional features or smart solutions
- **Documentation:** A clear README explaining your approach
- **Deployment**: If you are building a web application, it should be hosted and accessible
- **Architecture:** A well-thought-out architecture that is scalable and maintainable

If you are building a mobile app, you don’t have to publish it to the Play Store or App Store, but it should be easy to run in Android Studio/Xcode, and your backend should be hosted.

The deployment does not have to be anything fancy - as long as we can enter a URL in a browser and access your project, it’s sufficient.

## Submitting Your Project 🚀

Create a private GitHub repository and upload your code. Grant read access to [gendev@check24.de](mailto:gendev@check24.de).
When submitting your application, include the repository link.

**Your submission should include:**

- Your working code (backend and frontend/mobile app)
- A README.md explaining your approach
- A demo of the project:
  - If it’s a web app, a link to a hosted version
  - If it’s a mobile app, it should be runnable locally, with the app connecting to a hosted backend

Good luck and have fun coding! 😎

## Questions? ❓

If you have any questions, contact [gendev@check24.de](mailto:gendev@check24.de).
