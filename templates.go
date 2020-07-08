package main

const indexTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <link rel="stylesheet" href="https://unpkg.com/@picocss/pico@latest/css/pico.min.css">
    <meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>{{ .Title }}</title>
  </head>
<body>
  <nav>
    <ul>
	  <li><strong><a href="/">rss2twtxt</a></strong></li>
	  <li><a href="/feeds">Feeds</a></li>
    </ul>
  </nav>
  <main class="container">
	<hgroup>
	  <h2>rss2twtxt</h2>
	  <h3>RSS to twtxt</h3>
	</hgroup>
	<p>
	  rss2twtxt is a command-line tool and web app that processes RSS feeds
	nto <a href="https://twtxt.readthedocs.io/en/stable/index.html">twtxt</a>
	eeds for consumption by <i>twtxt</i> clients.
	</p>
	<form action="/" method="POST">
      <div class="grid">
		<label for="name">
		  Name:
		   <input type="text" id="name", name="name" placeholder="Feed's name" required>
		</label>

		<label for="url">
		  URL:
          <input type="url" id="url" name="url" placeholder="URL for the feed's RSS" required>
		</label>
      </div>
      <button type="submit">Add</button>
    </form>
  </main>
</body>
</html>
`

const feedsTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <link rel="stylesheet" href="https://unpkg.com/@picocss/pico@latest/css/pico.min.css">
    <meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>{{ .Title }}</title>
  </head>
<body>
  <nav>
    <ul>
	  <li><strong><a href="/">rss2twtxt</a></strong></li>
	  <li><a href="/feeds">Feeds</a></li>
    </ul>
  </nav>
  <main class="container">
	<ul>
	  {{ range .Feeds }}
		<li><a href="{{ .URL }}">{{ .Name }}</a>&nbsp;<small>({{ .LastModified }})</small></li>
	  {{ end }}
	</ul>
  </main>
</body>
</html>
`

const messageTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <link rel="stylesheet" href="https://unpkg.com/@picocss/pico@latest/css/pico.min.css">
    <meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>{{ .Title }}</title>
  </head>
<body>
  <nav>
    <ul>
	  <li><strong><a href="/">rss2twtxt</a></strong></li>
	  <li><a href="/feeds">Feeds</a></li>
    </ul>
  </nav>
  <main class="container">
	<p>{{ .Message }}</p>
  </main>
</body>
</html>
`
