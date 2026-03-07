# AuthApi

All URIs are relative to *http://e.alenalex/me/api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**authGoogleCallbackPost**](AuthApi.md#authgooglecallbackpost) | **POST** /auth/google/callback | Google OAuth callback |
| [**authGoogleLoginGet**](AuthApi.md#authgoogleloginget) | **GET** /auth/google/login | Initiate Google OAuth login |



## authGoogleCallbackPost

> HandlersGoogleCallbackResponse authGoogleCallbackPost(handlersGoogleCallbackRequest)

Google OAuth callback

Validates Google OAuth code and state, creates or retrieves user, sets access_token and refresh_token HttpOnly cookies.

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { AuthGoogleCallbackPostRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // HandlersGoogleCallbackRequest | Google OAuth callback request
    handlersGoogleCallbackRequest: ...,
  } satisfies AuthGoogleCallbackPostRequest;

  try {
    const data = await api.authGoogleCallbackPost(body);
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
| **handlersGoogleCallbackRequest** | [HandlersGoogleCallbackRequest](HandlersGoogleCallbackRequest.md) | Google OAuth callback request | |

### Return type

[**HandlersGoogleCallbackResponse**](HandlersGoogleCallbackResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | User authenticated successfully |  -  |
| **400** | Invalid request or state |  -  |
| **401** | OAuth validation failed |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## authGoogleLoginGet

> HandlersGoogleLoginResponse authGoogleLoginGet()

Initiate Google OAuth login

Generates OAuth state, stores it in HttpOnly cookie, and returns Google authorization URL.

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { AuthGoogleLoginGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  try {
    const data = await api.authGoogleLoginGet();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**HandlersGoogleLoginResponse**](HandlersGoogleLoginResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Returns Google OAuth URL |  -  |
| **500** | Failed to generate state |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

