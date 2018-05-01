
# Authentications 

## API key in HTTP header

``X-Api-Key: XXXXXXXXXXXXXXXXXXXX``


## User auth  & cookie

TODO


# Resources

## Photo

### POST /api/v1/photo

query params: 
- name -> model.Photo.Name*

body: photo

response:
- HTTP status
- model.Photo (JSON) 
    
### PUT /api/v1/photo/:id

body: model.Photo (JSON)

response: 204 (if success)
        
### GET /api/v1/photo/:id

response: -> JSON model.Photo

### DELETE /api/v1/photo/:id

response: 204 (if success)

### GET /api/v1/photo/search

for now return all photos

response: []model.Photo (JSON) 
