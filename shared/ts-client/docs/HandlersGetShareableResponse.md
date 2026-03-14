
# HandlersGetShareableResponse


## Properties

Name | Type
------------ | -------------
`activeFrom` | string
`createdAt` | string
`data` | string
`expiryAt` | string
`files` | [Array&lt;HandlersShareableFileResponse&gt;](HandlersShareableFileResponse.md)
`id` | string
`name` | string
`options` | { [key: string]: string; }
`type` | string
`userEmail` | string
`userId` | string

## Example

```typescript
import type { HandlersGetShareableResponse } from ''

// TODO: Update the object below with actual values
const example = {
  "activeFrom": null,
  "createdAt": null,
  "data": null,
  "expiryAt": null,
  "files": null,
  "id": null,
  "name": null,
  "options": null,
  "type": null,
  "userEmail": null,
  "userId": null,
} satisfies HandlersGetShareableResponse

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as HandlersGetShareableResponse
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


