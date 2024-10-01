# oauth-email-lists

This application leverages the power of third-party OAuth providers, allowing users to build email lists from returned OAuth data, and seamlessly integrates with various email service providers.

## Features

- Create and manage email lists
- Generate campaign links for easy subscriber sign-up flow
- Integrate with Google and Discord OAuth Providers, with more coming soon
- Redirect users to a specified URL after subscription
- Post subscriber data to third-party applications on sign-up (currently supports output to Aweber, Brevo, Resend, and Telegram)

# How it Works

The application is essentially an opt-in funnel.

Traffic enters the funnel at the Campaign URL, is taken to the Application, and then redirected to the specified OAuth Provider (Google, Discord, etc). At this point the visitor has the option to authorize the Application. If they agree, they are redirected back to the Application, and then out to the final URL.

In addition, a callback is triggered containing the subscriber info (email address, name, etc). This info is saved to the database, and (optionally) sent to any 3rd-party Outputs specified in the Campaign.

```base
                                                       -------> (8) Subscriber info is sent to 3rd-party output(s) (optional)
                                                      |    
                                                      |          ______
                                                      |         |      |       
                                                      |-------> |  DB  | (7) Subscriber info is saved to DB
                                                      |         |______|
                                             _________|
                                            |         | ------> (6) Visitor is redirected to final URL
    (1) Visitor clicks on Campaign URL ---> |   App   |
                                            |_________| <-------------------------------------------
                                              |     ^                                               |
  (2) Visitor is redirected to OAuth Provider |     |                                               |
                                              |     | (5) Visitor is redirected back to App         |
                                              ⌄     |                                               |
                                             __________                                             |
                                            |          |                                            |
                                            |  OAuth   |                                            |
                                            | Provider | -------------------------------------------
                                            | (Google) |    (4) Subscriber info is sent to App via callback
                                            |__________|

                                     (3) Visitor authorizes App
```

## Quickstart

To get the project running locally make sure you have [Go](https://go.dev/doc/install) and [Docker](https://docs.docker.com/engine/install) installed, then follow these steps:

1. Clone the repository:

```bash
git clone https://github.com/EricFrancis12/oauth-email-lists.git
```

2. Enter project directory:

```bash   
cd oauth-email-lists
```

3. Install dependencies:

```bash
go mod download
```

4. Run the following command to create a boilerplate `.env` file at the project root:

```bash
make create-env
```

5. Spin up a local Postgres instance:

```bash
docker compose up -d
```

6. Run the application:

```bash
make run
```

The application should now be running at http://localhost:6009 by default.

## Creating new Users

The Root User has the ability to create new users. Log in as the root user by visiting `/login`, and entering the `ROOT_USERNAME` and `ROOT_PASSWORD` from the `.env` file into the form.

Then, create a new user by sending a `POST` request to `/users`:

```bash
curl -X POST "http://localhost:6009/users" \
     -H "Content-Type: application/json" \
     -d '{
           "name": "Jim Bob",
           "password": "abcdefgh"
        }'
```

To verify the User was created successfully, make a `GET` request to `/users`. The response should include the newly-created User:

```json
{
    "success": true,
    "data": [
        {
            "id": "sdq0e64g-5lq2-467m-9xs6-s0fp4945xlgf",
            "name": "Jim Bob",
            "hashedPassword": "$2a$10$k4r2SwdJWa62ql/s4J4qcpcbVqt5o.JEBqWpUAgDPFIU/6cJP.iu8i",
            "createdAt": "2024-08-22T20:26:06.874752Z",
            "updatedAt": "2024-08-22T20:26:06.874752Z"
        }
    ]
}
```

## Email Lists

A User may have many Email Lists associated with them. To create a new Email List, make a `POST` request to `/email-lists`, making sure to reference the User ID that it should be attached to.

```bash
curl -X POST "http://localhost:6009/email-lists" \
     -H "Content-Type: application/json" \
     -d '{
           "name": "My First Email List",
           "userId": "sdq0e64g-5lq2-467m-9xs6-s0fp4945xlgf"
        }'
```

Verify the Email List was created successfully by making a `GET` request to `/email-lists`. You should see the newly-created Email List:

```json
{
    "success": true,
    "data": [
        {
            "id": "9ealnr84-lap9-4194-sko9-7a2aq4571nr6",
            "userId": "sdq0e64g-5lq2-467m-9xs6-s0fp4945xlgf",
            "name": "My First Email List",
            "createdAt": "2024-08-22T20:27:21.874752Z",
            "updatedAt": "2024-08-22T20:27:21.874752Z"
        }
    ]
}
```

## Outputs

Outputs are third-party applications that can be interacted with when a new subscriber is added to an Email List. More outputs will be added soon. Currently supported outputs include:

- [Aweber](https://aweber.com)
- [Brevo](https://brevo.com)
- [Resend](https://resend.com)
- [Telegram](https://telegram.org)

### Aweber
Coming soon

### Brevo
Coming soon

### Resend
Coming soon

### Telegram
Coming soon

## OAuth Providers

Users must specify an OAuth Provider when creating a campaign. More providers will be added soon. Currently supported providers include:

- Google
- Discord

### Google
Coming soon

### Discord
Coming soon

## Creating a Campaign

Any User may create a Campaign by making a `POST` request to the `/c` endpoint:

```bash
curl -X POST "http://localhost:6009/c" \
     -H "Content-Type: application/json" \
     -d '{
           "emailListId": "9ealnr84-lap9-4194-sko9-7a2aq4571nr6",
           "providerName": "Google",
           "outputIds": [
              ""
           ],
           "redirectUrl": "https://bing.com?src=my-redirect-url"
        }'
```

Response:

```json
{
	"success": true,
	"data": "http://localhost:6009/c?c=eyJhbGciOiJIUzI1NJ9.eyJkExLWIwNTktiIsIsnR5cCI6IkpXVCNTik2YTgUzNGFkiYi1JVxMDAyiNiVcVcdTAwMjYvlXHUw0ElMkYlMkwr5wnLmNvTAwMjYlX4ZyKKbSgUzRnNyYyUzRG15LXJlZGlyZWNYeXRhIjoiMWhjduN1MDAyNiVcdTAwMjYlXMDI2JVx1MDAyNiZhZGQxZDEtZxGVmOC00ZD0LXVybCJ9.PyCc0f"
}
```

This is the Campaign URL you would use as the entry-point into the funnel.
