package main

const indexTemplate = `
<!DOCTYPE html>
<html lang="zh">
  <head>
    <link rel="stylesheet" href="https://unpkg.com/@picocss/pico@latest/css/pico.min.css">
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>rss2twt :: {{ .Title }}</title>
  </head>
<body>
  <nav class="container-fluid">
    <ul>
      <li><strong><a href="/">rss2twt 中文版</a></strong></li>
      <li><a href="/feeds">Feeds</a></li>
    </ul>
  </nav>
  <main class="container">
    <article class="grid">
      <div>
        <hgroup>
          <h2>rss2twt 中文版</h2>
          <footer>RSS/Atom 转换到 Twtxt Feed</footer>
        </hgroup>
        <p>
		<b>注意：</b> 请添加 <b>中文</b> Feed 源，此站点主要为中文用户服务
		<p>
		</p>
		rss2twt是一个命令行工具和Web应用程序，可以将 RSS/Atom Feed 转换为 <a href="https://twtxt.readthedocs.io/en/stable/index.html">Twtxt</a> Feed，以供 Twtxt 客户端（例如 <a href="https://www.twtxt.cc">twtxt.cc</a> 和 <a href="https://www.twtxt.net">twtxt.net</a>）使用。
        </p>
        <p>
		您可以在这里任意添加新的 Feed 源，只需将网站的 URL 填入下面的文本输入框内，系统就会自动探索 RSS/Atom 源，如果 Feed 源有效，则将其添加到 <a href="/feeds">Feeds</a> 列表中。
        </p>
        <p>
		您可以使用自己喜欢的 <i>Twtxt</i> 客户端订阅 <a href="/feeds">Feed 源</a>（<i>我个人喜欢使用 <a href="https://github.com/quite/twet">twet</a></i>）。
		</p>
        <div class="container-fluid">
          <form action="/" method="POST">
            <input type="url" id="url" name="url" placeholder="Feed 源地址" required>
            <div><button type="submit">添加</button>
          </form>
        </div>
      </div>
    </article>
  </main>
  <footer class="container-fluid">
    <hr>
    <p>
      <small>
        Licensed under the <a href="https://github.com/twtpub/rss2twt/blob/master/LICENSE" class="secondary">MIT License</a><br>
      </small>
    </p>
  </footer>
</body>
</html>
`

const feedsTemplate = `
<!DOCTYPE html>
<html lang="zh">
  <head>
    <link rel="stylesheet" href="https://unpkg.com/@picocss/pico@latest/css/pico.min.css">
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>rss2twt :: {{ .Title }}</title>
  </head>
<body>
  <nav class="container-fluid">
    <ul>
      <li><strong><a href="/">rss2twt 中文版</a></strong></li>
      <li><a href="/feeds">Feeds</a></li>
    </ul>
  </nav>
  <main class="container">
    <article class="grid">
      <div>
        <hgroup>
          <h2>Feed 源</h2>
          <footer>可用的 Twtxt Feeds</footer>
        </hgroup>
        {{ if .Feeds }}
          <ul>
            {{ range .Feeds }}
              <li><a href="{{ .URL }}">{{ .Name }}</a>&nbsp;<small>({{ .LastModified }})</small></li>
            {{ end }}
          </ul>
        {{ else }}
          <small>还没有可用的 Feed，请稍候再来！</small>
        {{ end }}
      </div>
    </article>
  </main>
  <footer class="container-fluid">
    <hr>
    <p>
      <small>
        Licensed under the <a href="https://github.com/twtpub/rss2twt/blob/master/LICENSE" class="secondary">MIT License</a><br>
      </small>
    </p>
  </footer>
</body>
</html>
`

const messageTemplate = `
<!DOCTYPE html>
<html lang="zh">
  <head>
    <link rel="stylesheet" href="https://unpkg.com/@picocss/pico@latest/css/pico.min.css">
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{{ .Title }}</title>
  </head>
<body>
  <nav class="container-fluid">
    <ul>
      <li><strong><a href="/">rss2twt 中文版</a></strong></li>
      <li><a href="/feeds">Feeds</a></li>
    </ul>
    <ul>
      <li>
        <a href="https://github.com/txtpub/rss2twt" class="contrast" aria-label="Pico GitHub repository">
          <svg aria-hidden="true" focusable="false" role="img" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 496 512" height="1rem">
            <path fill="currentColor" d="M165.9 397.4c0 2-2.3 3.6-5.2 3.6-3.3.3-5.6-1.3-5.6-3.6 0-2 2.3-3.6 5.2-3.6 3-.3 5.6 1.3 5.6 3.6zm-31.1-4.5c-.7 2 1.3 4.3 4.3 4.9 2.6 1 5.6 0 6.2-2s-1.3-4.3-4.3-5.2c-2.6-.7-5.5.3-6.2 2.3zm44.2-1.7c-2.9.7-4.9 2.6-4.6 4.9.3 2 2.9 3.3 5.9 2.6 2.9-.7 4.9-2.6 4.6-4.6-.3-1.9-3-3.2-5.9-2.9zM244.8 8C106.1 8 0 113.3 0 252c0 110.9 69.8 205.8 169.5 239.2 12.8 2.3 17.3-5.6 17.3-12.1 0-6.2-.3-40.4-.3-61.4 0 0-70 15-84.7-29.8 0 0-11.4-29.1-27.8-36.6 0 0-22.9-15.7 1.6-15.4 0 0 24.9 2 38.6 25.8 21.9 38.6 58.6 27.5 72.9 20.9 2.3-16 8.8-27.1 16-33.7-55.9-6.2-112.3-14.3-112.3-110.5 0-27.5 7.6-41.3 23.6-58.9-2.6-6.5-11.1-33.3 2.6-67.9 20.9-6.5 69 27 69 27 20-5.6 41.5-8.5 62.8-8.5s42.8 2.9 62.8 8.5c0 0 48.1-33.6 69-27 13.7 34.7 5.2 61.4 2.6 67.9 16 17.7 25.8 31.5 25.8 58.9 0 96.5-58.9 104.2-114.8 110.5 9.2 7.9 17 22.9 17 46.4 0 33.7-.3 75.4-.3 83.6 0 6.5 4.6 14.4 17.3 12.1C428.2 457.8 496 362.9 496 252 496 113.3 383.5 8 244.8 8zM97.2 352.9c-1.3 1-1 3.3.7 5.2 1.6 1.6 3.9 2.3 5.2 1 1.3-1 1-3.3-.7-5.2-1.6-1.6-3.9-2.3-5.2-1zm-10.8-8.1c-.7 1.3.3 2.9 2.3 3.9 1.6 1 3.6.7 4.3-.7.7-1.3-.3-2.9-2.3-3.9-2-.6-3.6-.3-4.3.7zm32.4 35.6c-1.6 1.3-1 4.3 1.3 6.2 2.3 2.3 5.2 2.6 6.5 1 1.3-1.3.7-4.3-1.3-6.2-2.2-2.3-5.2-2.6-6.5-1zm-11.4-14.7c-1.6 1-1.6 3.6 0 5.9 1.6 2.3 4.3 3.3 5.6 2.3 1.6-1.3 1.6-3.9 0-6.2-1.4-2.3-4-3.3-5.6-2z"></path>
          </svg>
        </a>
      </li>
    </ul>
  </nav>
  <main class="container">
    <article class="grid">
      <div>
        <p>{{ .Message }}</p>
      </div>
    </article>
  </main>
  <footer class="container-fluid">
    <hr>
    <p>
      <small>
        Licensed under the <a href="https://github.com/twtpub/rss2twt/blob/master/LICENSE" class="secondary">MIT License</a><br>
      </small>
    </p>
  </footer>
</body>
</html>
`
