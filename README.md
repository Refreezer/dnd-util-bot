## Simple telegram bot for group DnD games

```bash
        docker run --rm --name dnd-util-bot -d \
                    -v *persistent volume name*:*mount path* \
                    -e DND_UTIL_RATE_LIMIT_RPS=200 \
                    -e DND_UTIL_LONG_POLLING_TIMEOUT=60 \
                    -e DND_UTIL_BOT_NAME=dnd_util_bot \
                    -e DND_UTIL_TG_API_KEY=*your api key* \
                    -e DND_UTIL_DB_PATH=/var/data/dndUtil.db \
                    -e DND_UTIL_BOT_NAME=dnd_util_bot \
                    *docker image name*
```
