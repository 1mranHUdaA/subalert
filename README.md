# SUBALERT 🚨

**Subalert is an automation tool to get newly updated subdomains from your target.**

It is combined with Subfinder and discord to give alert on newly added subdomains. The alert will contain the newly added subdomain if it is **Alive/Dead**

# Use cases

`go run main.go -d hackerone.com` to set alert for single domain.

`go run main.go -f domain_list.txt` to set alert for given root domains.

Make sure you replace your Discord webhook in the `main.go`

It is recomended to use `tmux` or `screen` so newly added subdomains can be monitored after 24 hours.
