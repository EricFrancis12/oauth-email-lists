# OAuth Email Lists

This application leverages the power of third-party OAuth providers, allowing users to build email lists from returned OAuth data, and seamlessly integrates with various email service providers.

## Features

- Create and manage email lists
- Generate campaign links for easy subscriber sign-up flow
- Integrate with Google and Discord OAuth Providers, with more coming soon
- Redirect users to a specified URL after subscription
- Post subscriber data to third-party applications on sign-up (currently supports output to Aweber, Brevo, Resend, and Telegram)

## How it Works

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
To integrate with AWeber, simply sign up for an account at https://aweber.com, and create an email list. Then navigate to `List Options` -> `List Settings` and get your List ID (see image below).

<img src="https://github.com/user-attachments/assets/c8368ff7-4109-4fbe-944c-60f72422bed2"/>

Now make a `POST` request to `/outputs`:

```bash
curl -X POST "http://localhost:6009/outputs" \
     -H "Content-Type: application/json" \
     -d '{
           "userId": "sdq0e64g-5lq2-467m-9xs6-s0fp4945xlgf",
           "outputName": "aweber",
           "listId": "awlist8157462",
           "param1": "[optional AWeber ad tracking]"
        }'
```

### Brevo
To integrate with Brevo, sign up at https://brevo.com. Navigate to `Contacts` -> `Lists` in the Brevo dashboard (see image below).

<img src="https://github.com/user-attachments/assets/560a5b03-9dd1-4d2e-be41-32e1f92a5433" />

Brevo uses numeric list IDs unique to each account, so the List ID that we want is `2`.

Next, go to https://app.brevo.com/settings/keys/api and create a new API Key. Then add your API Key to the `.env` file for `BREVO_API_KEY`.

Now we can make a `POST` request to create a new Brevo Output:

```bash
curl -X POST "http://localhost:6009/outputs" \
     -H "Content-Type: application/json" \
     -d '{
           "userId": "sdq0e64g-5lq2-467m-9xs6-s0fp4945xlgf",
           "outputName": "brevo",
           "listId": "2",
           "param1": "[optional EXT_ID Contact Attribute ]"
        }'
```

### Resend
To integrate with Resend, sign up at https://resend.com.

Resend uses an Audience ID to uniquely identify Audiences on their platform. Navigate to the `Audiences` tab on their dashboard, and create a new Audience if needed, and grab your Audience ID (see image below).

<img src="https://github.com/user-attachments/assets/054b3822-917f-49ab-85ff-9aba30180f5d" />

Next, go to the `API Keys` tab and create a new API Key. Then add your API Key to the `.env` file for `RESEND_API_KEY`.

Finally, make a `POST` request to `/outputs` to create a new Resend Output:

```bash
curl -X POST "http://localhost:6009/outputs" \
     -H "Content-Type: application/json" \
     -d '{
           "userId": "sdq0e64g-5lq2-467m-9xs6-s0fp4945xlgf",
           "outputName": "resend",
           "listId": "e0a3e864-da54-49"
        }'
```

### Telegram
To integrate with Telegram, you will need to obtain a Telegram Bot ID, as well as the Chat ID of where the messages should be sent.

This video explains how to create a Telegram Bot and get the Bot ID: https://www.youtube.com/watch?v=Qe16589RNNY

This video explains how to get your Telegram Chat ID: https://www.youtube.com/watch?v=uXhFsScozyY

Add your Bot ID to the `.env` file for `TELEGRAM_BOT_ID`.

Now you need to define the message content. You can substitute in values for `subscriber.name` and `subscriber.emailAddr` by using `{{name}}` and `{{emailAddr}}` respectively. For example, if this is the message content:

```
"A new subscriber was just added. Their name is {{name}}, and their email address is {{emailAddr}}."
```

And a new subscriber named Tom Jones, with an email address of tomjones@domain.com subscribes, the message will be evaluated to:

```
"A new subscriber was just added. Their name is Tom Jones, and their email address is tomjones@domain.com."
```

Now we can make a `POST` request to `/outputs` to create a new Telegram Output:

```bash
curl -X POST "http://localhost:6009/outputs" \
     -H "Content-Type: application/json" \
     -d '{
           "userId": "sdq0e64g-5lq2-467m-9xs6-s0fp4945xlgf",
           "outputName": "telegram",
           "listId": "[your-chat-id]",
           "param1": "[your-message-content]"
        }'
```

## OAuth Providers

Users must specify an OAuth Provider when creating a campaign. More providers will be added soon. Currently supported providers include:

- Google
- Discord

### Google
To integrate with Google as an OAuth Provider, visit https://console.cloud.google.com and create a new project. Then navigate to https://console.cloud.google.com/apis/credentials and create your credentials. After that you will have an Client ID and Client Secret for your project. 

Now add your Client ID and Client Secret to the `.env` file for `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` respectively.

Also, you will need to set Authorized JavaScript origins and Authorized redirect URIs in your console for wherever you plan to run the app (see image below).

<img src="https://github.com/user-attachments/assets/87ad0f9c-3172-49e7-ab3f-2935928f7e74" />

### Discord
To integrate with Discord as an OAuth Provider, create a Discord developer account if you don't already have one, then navigate to https://discord.com/developers/applications and create a new application.

Grab your Client ID and Client Secret, and add them to the `.env` file for `DISCORD_CLIENT_ID` and `DISCORD_CLIENT_SECRET` respectively.

Also, you will need to add Redirects in your dashboard for wherever you plan to run the app (see image below).

<img src="https://github.com/user-attachments/assets/e1c101d4-2b50-45e2-be11-5ceb17c937a1" />

## Creating a Campaign

Any User may create a Campaign by making a `POST` request to the `/c` endpoint:

```bash
curl -X POST "http://localhost:6009/c" \
     -H "Content-Type: application/json" \
     -d '{
           "emailListId": "9ealnr84-lap9-4194-sko9-7a2aq4571nr6",
           "providerName": "Google",
           "outputIds": [
              "[your-first-output-id]",
              "[your-second-output-id]"              
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

This is the Campaign URL you would use as the entry-point to the funnel.
