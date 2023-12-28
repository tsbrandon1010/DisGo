# DisGo Bot
---
DisGo Bot is a simple Discord bot written in GoLang. The primary purpose of this bot is to allow for the streaming of audio in Discord voice channels. This is enabled by [DCA](https://github.com/jonas747/dca), [FFmpeg](https://ffmpeg.org/)
and [yt-dlp](https://github.com/yt-dlp/yt-dlp).

## Setting Up The Bot:

There are two primary methods for running the bot:
1. Building and running the bot using Docker.
2. Running/building the binaries yourself.

First, you will need to set up a Discord bot app in the Discord Developer Portal, and invite the bot to your Discord. Resources to do this can be found [here](https://discord.com/developers/docs/getting-started) (Step 1. Creating an app).

### Using Docker:

This is the simplest method, and only requires you to clone the repository, and build/run the Docker image (assuming
that you have Docker installed).

After downloading or cloning the repository, create an .env file which contains the following fields:

![TOKEN="YOUR DISCORD BOT TOKEN HERE",
APP_ID="YOUR DISCORD APP'S ID", 
STATUS_URL="AN OPTIONAL URL FOR THE BOT'S STATUS
](https://github.com/tsbrandon1010/DisGo/assets/15933213/13036cad-1dd1-4b48-afed-60932c4fca52)

After doing so, you can build the Docker image using the following command (while in the directory of the cloned repository): ```docker build -t disgo-bot .```.
This command creates a Docker image named "disgo-bot".

Finally, you can run the bot by entering the following command: ```docker -d --name disgo-bot disgo-bot```.

This will create a Docker container which is detached (running in the background), named "disgo-bot", which utilizes the image that we created in the previous step. To stop the bot, you can run the command ```docker stop disgo-bot```,
and optionally ```docker remove disgo-bot``` to delete the container.

### Building & Running The Binaries Yourself:
This method requires that you have the following dependencies installed:
* [FFmpeg](https://ffmpeg.org/)
* [yt-dlp](https://github.com/yt-dlp/yt-dlp)
* [Go](https://go.dev/doc/install)

Create an .env file which contains the following fields:

![TOKEN="YOUR DISCORD BOT TOKEN HERE",
APP_ID="YOUR DISCORD APP'S ID", 
STATUS_URL="AN OPTIONAL URL FOR THE BOT'S STATUS
](https://github.com/tsbrandon1010/DisGo/assets/15933213/13036cad-1dd1-4b48-afed-60932c4fca52)

After downloading or cloning the repository, and creating a .env file, run the following command while in the directory of the cloned repository: ```go build```. After doing so, there should be a binary named "main" in the directory which can be run to start the bot.
