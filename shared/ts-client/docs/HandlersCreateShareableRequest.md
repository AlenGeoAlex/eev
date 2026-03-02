
# HandlersCreateShareableRequest


## Properties

Name | Type
------------ | -------------
`allowedEmails` | Array&lt;string&gt;
`data` | string
`files` | [Array&lt;HandlersShareableFileRequest&gt;](HandlersShareableFileRequest.md)
`name` | string
`notificationParams` | [HandlersCreateShareableRequestNotificationParams](HandlersCreateShareableRequestNotificationParams.md)
`options` | [HandlersCreateShareableRequestParams](HandlersCreateShareableRequestParams.md)
`timeParams` | [HandlersCreateShareableRequestTimeParams](HandlersCreateShareableRequestTimeParams.md)
`type` | [ServicesShareableType](ServicesShareableType.md)

## Example

```typescript
import type { HandlersCreateShareableRequest } from ''

// TODO: Update the object below with actual values
const example = {
  "allowedEmails": null,
  "data": null,
  "files": null,
  "name": null,
  "notificationParams": null,
  "options": null,
  "timeParams": null,
  "type": null,
} satisfies HandlersCreateShareableRequest

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as HandlersCreateShareableRequest
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


