
# ServicesShareableCode


## Properties

Name | Type
------------ | -------------
`createdAt` | string
`expiryAt` | string
`id` | string
`name` | string
`options` | { [key: string]: string; }
`shareableData` | string
`shareableType` | string
`sourceIp` | string
`userEmail` | string
`userID` | string

## Example

```typescript
import type { ServicesShareableCode } from ''

// TODO: Update the object below with actual values
const example = {
  "createdAt": null,
  "expiryAt": null,
  "id": null,
  "name": null,
  "options": null,
  "shareableData": null,
  "shareableType": null,
  "sourceIp": null,
  "userEmail": null,
  "userID": null,
} satisfies ServicesShareableCode

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ServicesShareableCode
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


