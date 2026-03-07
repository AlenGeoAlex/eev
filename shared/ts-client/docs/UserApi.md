# UserApi

All URIs are relative to *http://e.alenalex/me/api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**meEmailHistoryGet**](UserApi.md#meemailhistoryget) | **GET** /me/email-history | Get the past email targets of the user |
| [**meGet**](UserApi.md#meget) | **GET** /me | Get current logged-in user info |



## meEmailHistoryGet

> HandlersTargetUserEmailResponse meEmailHistoryGet(search)

Get the past email targets of the user

Returns the past email targets of the user

### Example

```ts
import {
  Configuration,
  UserApi,
} from '';
import type { MeEmailHistoryGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: CookieAuth
    apiKey: "YOUR API KEY",
  });
  const api = new UserApi(config);

  const body = {
    // string | Search email substring (optional)
    search: search_example,
  } satisfies MeEmailHistoryGetRequest;

  try {
    const data = await api.meEmailHistoryGet(body);
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
| **search** | `string` | Search email substring | [Optional] [Defaults to `undefined`] |

### Return type

[**HandlersTargetUserEmailResponse**](HandlersTargetUserEmailResponse.md)

### Authorization

[CookieAuth](../README.md#CookieAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Successfully retrieved user history |  -  |
| **401** | Unauthorized - missing or invalid token |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## meGet

> HandlersMeResponse meGet()

Get current logged-in user info

Returns the currently authenticated user based on the access_token HttpOnly cookie.

### Example

```ts
import {
  Configuration,
  UserApi,
} from '';
import type { MeGetRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: CookieAuth
    apiKey: "YOUR API KEY",
  });
  const api = new UserApi(config);

  try {
    const data = await api.meGet();
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

[**HandlersMeResponse**](HandlersMeResponse.md)

### Authorization

[CookieAuth](../README.md#CookieAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Successfully retrieved user |  -  |
| **401** | Unauthorized - missing or invalid token |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

