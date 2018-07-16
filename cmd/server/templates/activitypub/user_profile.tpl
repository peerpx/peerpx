{
    "@context": [
        "https://www.w3.org/ns/activitystreams",
        "https://w3id.org/security/v1",
        {
            "manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
            "sensitive": "as:sensitive",
            "movedTo": "as:movedTo",
            "Hashtag": "as:Hashtag"
        }
    ],
    "id": "https://{{.BaseURL}}/users/{{.UserName}}",
    "type": "Person",
    "following": "https://{{.BaseURL}}/users/{{.UserName}}/following",
    "followers": "https://{{.BaseURL}}/users/{{.UserName}}/followers",
    "inbox": "https://{{.BaseURL}}/users/{{.UserName}}/inbox",
    "outbox": "https://{{.BaseURL}}/users/{{.UserName}}/outbox",
    "featured": "https://{{.BaseURL}}/users/{{.UserName}}/collections/featured",
    "preferredUsername": "{{.UserName}}",
    "name": "{{.UserName}}",
    "summary": "{{.Summary}}",
    "url": "https://{{.BaseURL}}/@{{.UserName}}",
    "manuallyApprovesFollowers": false,
    "publicKey": {
        "id": "https://{{.BaseURL}}/users/{{.UserName}}#main-key",
        "owner": "https://{{.BaseURL}}/users/{{.UserName}}",
        "publicKeyPem": "{{.PubKey}}"
    },
    "tag": [],
    "attachment": [],
    "endpoints": {
        "sharedInbox": "https://{{.BaseURL}}/inbox"
    },
    "icon": {
        "type": "Image",
        "mediaType": "image/png",
        "url": ""
    },
    "image": {
        "type": "Image",
        "mediaType": "image/jpeg",
        "url": ""
    }
}