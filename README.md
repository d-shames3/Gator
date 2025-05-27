# GatorCLI

Gator is a command line RSS feed aggre-GATOR. Users can subscribe to their favorite RSS feeds and navigate to posts that they want to read, all from their terminal. 

### Prerequisites
You'll need to install the following before you get started
* [Go](https://go.dev/doc/install)
* [Postgres](https://www.postgresql.org/download/)

### Install Gator
```bash
go install github.com/d-shames3/gatorcli@latest
```

### Configure Database
1. Ensure postgres was installed properly by running: 
```bash
psql --version
```
2. Start Postgres in the background using Homebrew if on mac:
```bash
brew services start postgresql@15
```
3. Enter the psql shell:
```bash
psql postgres
```
4. Create a new database called gator:
```SQL
CREATE DATABASE gator;
```
5. Connect to the gator database:
```bash
\c gator
```
6. Type exit to leave the psql shell. 

7. Determine your connection string for the database. The format for mac is:
```bash
postgres://your-user-name-here:@localhost:5432/gator
```
Test that your connection string works by running (swap in your connection string):
```bash
psql postgres://your-user-name-here:@localhost:5432/gator 
```

8. Create a file in your home directory called `.gatorconfig.json`. The file should have the following contents:
```JSON
{
  "db_url": "postgres://your-user-name-here:@localhost:5432/gator"
}
```
Make sure to swap out the template connection string for your own. 

9. Run the [goose](https://github.com/pressly/goose) migrations to get your database set up with the correct tables:
```bash
cd gatorcli/sql/schema
goose postgres postgres://your-user-name-here:@localhost:5432/gator up
```
Make sure your database now has the the relevant tables by running:
```bash
psql postgres://your-user-name-here:@localhost:5432/gator
\dt
```
If you see tables such as `users`, `feeds`, etc., you are ready to use the tool. 

### Usage
GatorCLI allows users to execute the following commands:

addfeed * agg * browse * feeds * follow *  following * login * register * reset * users * unfollow

For full usage, a user will have to first register. 

#### addfeed
Subscribes a user to an RSS feed. 

Required args: feed name, url

Example:
```bash
gator addfeed "PostHog" "https://newsletter.posthog.com/feed"
```

#### agg
Background service that aggregates fetching RSS feeds and stores post information in the gator database. Best used in a separate terminal that you can leave running in the background. 

Required args: time between requests (formatted as 10s, 30m, 100h, etc.). NOTE: do not DOS sites. Add a substantial backoff period. 

Example:
```bash
gator agg 24h
```

Execute `ctrl-C` to kill the `agg` service

#### browse
Prints the most recent posts from feeds you are following to your terminal. Defaults to 2 posts, but you can specify how many you want.

Optional args: number of posts (default is 2)

Example:
```bash
gator browse 10
```

#### feeds
Prints existing feeds that you can follow to the terminal. 

Example:
```bash
gator feeds
```

#### follow
Sets up user to follow a given feed. Any feed the user adds themselves will be auto-followed. 

Required args: feed url (Users can get a url by running `gator feeds`)

Example:
```bash
gator follow "https://newsletter.posthog.com/feed"
```

#### following
Prints feeds the user is currently following to the terminal. 

Example:
```bash
gator following
```

#### login
Logs a registered user into gator. Most functionality is restricted to a logged-in user.

Required args: username

Example:
```bash
gator login john-doe
```

#### register
Registers and logs in as a new user. Most functionality is resticted to registered/logged in users.

Required args: username

Example:
```bash
gator register john-doe
```

#### reset
WARNING: DESTRUCTIVE! Wipes your database of all data by deleting all users, which cascades deletes of all related data in the gator db. 

Example:
```bash
gator reset
```

#### unfollow
Unsubscribes a user from an RSS feed. 

Required args: feed url

Example:
```bash
gator unfollow "https://newsletter.posthog.com/feed" 
```

#### users
Lists all registered gator users and indicates who is currently logged in.

Example:
```bash
gator users
```
