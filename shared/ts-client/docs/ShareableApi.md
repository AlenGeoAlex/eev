# ShareableApi

All URIs are relative to *http://https://e.alenalex/me/api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**shareCodeGet**](ShareableApi.md#sharecodeget) | **GET** /share/{code} | Get shareable by code |
| [**sharePost**](ShareableApi.md#sharepost) | **POST** /share | Create a new shareable |



## shareCodeGet

> ServicesShareableCode shareCodeGet(code)

Get shareable by code

Retrieves public shareable information using its unique code.

### Example

```ts
import {
  Configuration,
  ShareableApi,
} from '';
import type { ShareCodeGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ShareableApi();

  const body = {
    // string | Shareable code
    code: code_example,
  } satisfies ShareCodeGetRequest;

  try {
    const data = await api.shareCodeGet(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **code** | `string` | Shareable code | [Defaults to `undefined`] |

### Return type

[**ServicesShareableCode**](ServicesShareableCode.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Shareable info |  -  |
| **400** | Invalid code |  -  |
| **404** | Shareable not found |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## sharePost

> HandlersCreateShareableResponse sharePost(handlersCreateShareableRequest)

Create a new shareable

Creates a new shareable resource (text, url, or file). Requires authentication via access_token cookie.

### Example

```ts
import {
  Configuration,
  ShareableApi,
} from '';
import type { SharePostRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: CookieAuth
    apiKey: "YOUR API KEY",
  });
  const api = new ShareableApi(config);

  const body = {
    // HandlersCreateShareableRequest | Create shareable request
    handlersCreateShareableRequest: ...,
  } satisfies SharePostRequest;

  try {
    const data = await api.sharePost(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **handlersCreateShareableRequest** | [HandlersCreateShareableRequest](HandlersCreateShareableRequest.md) | Create shareable request | |

### Return type

[**HandlersCreateShareableResponse**](HandlersCreateShareableResponse.md)

### Authorization

[CookieAuth](../README.md#CookieAuth)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Shareable created |  -  |
| **202** | Shareable created - awaiting encrypted data |  -  |
| **400** | Invalid request |  -  |
| **401** | Unauthorized |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

