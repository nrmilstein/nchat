# nchat

nchat is a chat api that allows users to send messages to each other.



## API:

### Register a new user

`POST /users`

Parameters:

| Name     | Type   | In   | Description                                           |
| -------- | ------ | ---- | ----------------------------------------------------- |
| email    | string | body | The email of the user you would like to register.     |
| password | string | body | The password of the user you would like to register.  |
| name     | string | body | The name of the user you would like to sign register. |

Sample response:

`Status: 201 Created`

```json
{
    "success": "user added",
    "user": {
        "created": "2020-09-12T17:12:45.170484+02:00",
        "email": "user@example.com",
        "id": 14,
        "name:": "User McUserface"
    }
}
```



### Generate an authentication key

`POST /authenticate`

Parameters:

| Name     | Type   | In   | Description                                              |
| -------- | ------ | ---- | -------------------------------------------------------- |
| email    | string | body | The email of the user you would like to authenticate.    |
| password | string | body | The password of the user you would like to authenticate. |

Sample response:

`Status: 201 Created`

```json
{
    "data": {
        "authKey": "Ozo1SNVJ7vxh0jkGIVfoN0dT",
        "user": {
            "email": "user@example.com",
            "id": 14,
            "name": "User McUserface"
        }
    },
    "status": "success"
}
```



### Get the authenticated User's details

`GET /authenticate`

Parameters:

| Name      | Type   | In     | Description                            |
| --------- | ------ | ------ | -------------------------------------- |
| X-API-KEY | string | header | The API key of the authenticated user. |

Sample response:

`Status: 200 OK`

```json
{
    "data": {
        "user": {
            "email": "user@example.com",
            "id": 14,
            "name": "User McUserface"
        }
    },
    "status": "success"
}
```



### Get the user's conversations

`GET /conversations`

| Name      | Type   | In     | Description                            |
| --------- | ------ | ------ | -------------------------------------- |
| X-API-KEY | string | header | The API key of the authenticated user. |

Sample response:

`Status: 200 OK`

```json
{
    "data": {
        "conversations": [
            {
                "id": 26,
                "users": [
                    {
                        "email": "user@example.com",
                        "id": 14,
                        "name": "User McUserface"
                    },
                    {
                        "email": "user2@example.com",
                        "id": 12,
                        "name": "User McDeeds"
                    }
                ]
            }
        ]
    },
    "status": "success"
}
```



### Get a conversation

`GET /conversations/{conversationId}`

Parameters:

| Name           | Type   | In     | Description                                       |
| -------------- | ------ | ------ | ------------------------------------------------- |
| X-API-KEY      | string | header | The API key of the authenticated user.            |
| conversationId | int    | path   | The id of the conversation you would like to get. |

Sample response:

`Status: 200 OK`

```json
{
    "data": {
        "conversation": {
            "created": "2020-09-12T17:17:15.727428+02:00",
            "id": 26,
            "messages": [
                {
                    "body": "Hello, world",
                    "id": 20,
                    "sent": "2020-09-12T17:17:15.727428+02:00",
                    "userId": 14
                },
                {
                    "body": "Hi there",
                    "id": 21,
                    "sent": "2020-09-12T17:19:29.193389+02:00",
                    "userId": 12
                }
            ],
            "users": [
                {
                    "email": "user@example.com",
                    "id": 14,
                    "name": "User McUserface"
                },
                {
                    "email": "user2@example.com",
                    "id": 12,
                    "name": "User McDeeds"
                }
            ]
        }
    },
    "status": "success"
}
```



### Start a new conversation

`POST /conversations`

| Name      | Type   | In     | Description                                                  |
| --------- | ------ | ------ | ------------------------------------------------------------ |
| X-API-KEY | string | header | The API key of the authenticated user.                       |
| userId    | int    | body   | The id of the user you would like to start a conversation with. |
| message   | string | body   | The opening message you would like to send.                  |

Sample response:

`Status: 201 Created`

```json
{
    "data": {
        "conversation": {
            "created": "2020-09-12T17:17:15.727428+02:00",
            "id": 26,
            "messages": [
                {
                    "body": "Hello, world",
                    "id": 20,
                    "sent": "2020-09-12T17:17:15.727428+02:00",
                    "userId": 14
                }
            ]
        }
    },
    "status": "success"
}
```





### Send a message via an existing conversation

`POST /conversations/{conversationId}`

| Name           | Type   | In     | Description                                                  |
| -------------- | ------ | ------ | ------------------------------------------------------------ |
| X-API-KEY      | string | header | The API key of the authenticated user.                       |
| conversationId | int    | path   | The id of the conversation you would like to send a message via. |
| message        | string | body   | The message you would like to send.                          |

Sample response:

`201 Created`

```json
{
    "data": {
        "message": {
            "body": "Hi there",
            "conversationId": 26,
            "id": 21,
            "sent": "2020-09-12T17:19:29.193389+02:00",
            "userId": 12
        }
    },
    "status": "success"
}
```

