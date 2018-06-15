
# Authentications 

## API key in HTTP header

``X-Api-Key: XXXXXXXXXXXXXXXXXXXX``


## User auth  & cookie

TODO


# Resources

## Photo

### POST /api/v1/photo

Request: 
body: multipart
    - properties: entities.Photo
    - file: photo 
Response:
response.Data = entities.Photo

    
### PUT /api/v1/photo

body: model.Photo (JSON)

response code: 200

response body: Json {"code": int, "photo": models.Photo }

code: 
- 0: ok
- 1: Name is too long (max length: 255)
- 2: Camera is too long (max length: 255)
- 3: Lens is too long (max length: 255)
- 4: ShutterSpeed is too long (max length: 255)
- 5: Location is too long (max length: 255)
- 6: Latitude is out of range ( -90.00 < latitude < +90.00
- 7: Longitude is out of range (-180 < longitude < +180.00
- 8: Hey Marty "TakenAt" is in the future !
        
### GET /api/v1/photo/:id

response: -> JSON model.Photo

### DELETE /api/v1/photo/:id

response: 204 (if success)

### GET /api/v1/photo/search

for now return all photos

response: []model.Photo (JSON) 
