# Emojify API
[![CircleCI](https://circleci.com/gh/emojify-app/api.svg?style=svg)](https://circleci.com/gh/emojify-app/api)

API for Emojify application

## Endpoints

### / POST
Create a new emojify image

**Post body**  
URI path to an image to be Emojified

**Response Codes**
Bad Request - Post body is not a valid URI
Internal Server - Internal failure
Not Modified - The posted URI already exists in the cache
Created - New Emojify request has been created
