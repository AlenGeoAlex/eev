# UserApi

All URIs are relative to *http://https://e.alenalex/me/api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**meGet**](UserApi.md#meget) | **GET** /me | Get current logged-in user info |



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

