# Cloud Computing Assignment 1
Made by **Svein Are Danielsen**

Studentno.: **253067** 


## URL
<https://eco-byte-216121.appspot.com/igcinfo/>

## Requests
### GET [/api](https://eco-byte-216121.appspot.com/igcinfo/api)
#### Usage
Used to retrieve meta-data about the API
##### Result
```json
{
  "uptime": "<uptime>",
  "info": "<information>",
  "version": "<version>"
}
```

### GET [/api/igc](https://eco-byte-216121.appspot.com/igcinfo/api/igc)
#### Usage
Retrieve IDs of all entries
##### Result
```json
[
  "<id0>",
  "<id1>",
  "<id1>"
]
```

### GET [/api/igc/`<id>`](https://eco-byte-216121.appspot.com/igcinfo/api/igc/1)
#### Usage
Retrieve data about an entry, based on it's ID
##### Result
```json
{
  "h_date": "<h_date>",
  "pilot": "<pilot>",
  "glider": "<glider>",
  "glider_id": "<glider_id>",
  "track_length": "<track_length>"
}
```

### GET [/api/igc/`<id>`/`<field>`](https://eco-byte-216121.appspot.com/igcinfo/api/igc/1/pilot)
#### Usage
Retrieve a data-field about and enrty, based on it's ID and field-name
##### Result
```text
<value>
```

### POST [/api/igc](https://eco-byte-216121.appspot.com/igcinfo/api/igc)
#### Usage
Insert a new entry with an URL to a `.igc` file
##### Body
```json
{
  "url": "<url>"
}
```
##### Response
```json
{
  "id": "<id>"
}
```