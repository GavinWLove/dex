package main

import (
	"github.com/Masterminds/sprig/v3"
	"html/template"
	"log"
	"net/http"
)

var bizTmpl = template.Must(template.New("biz.html").Parse(`<html>
  <head>
    <style>
/* make pre wrap */
pre {
 white-space: pre-wrap;       /* css-3 */
 white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
 white-space: -pre-wrap;      /* Opera 4-6 */
 white-space: -o-pre-wrap;    /* Opera 7 */
 word-wrap: break-word;       /* Internet Explorer 5.5+ */
}
    </style>
  </head>
  <body>
	<p> biz page </p>
    <p> ID Token: <pre><code>{{ .IDToken }}</code></pre></p>
    <p> Access Token: <pre><code>{{ .AccessToken }}</code></pre></p>
    <p> Claims: <pre><code>{{ .Claims }}</code></pre></p>
  </body>
</html>`))





var indexTmpl = template.Must(template.New("index.html").Parse(`<html>
  <head>
    <style>
form  { display: table;      }
p     { display: table-row;  }
label { display: table-cell; }
input { display: table-cell; }
    </style>
  </head>
  <body>
    <form action="/login" method="post">
      <p>
        <label> Authenticate for: </label>
        <input type="text" name="cross_client" placeholder="list of client-ids">
      </p>
      <p>
        <label>Extra scopes: </label>
        <input type="text" name="extra_scopes" placeholder="list of scopes">
      </p>
      <p>
        <label>Connector ID: </label>
        <input type="text" name="connector_id" placeholder="connector id">
      </p>
      <p>
        <label>Request offline access: </label>
        <input type="checkbox" name="offline_access" value="yes" checked>
      </p>
      <p>
	    <input type="submit" value="Login">
      </p>
    </form>
  </body>
</html>`))

func renderIndex(w http.ResponseWriter) {
	renderTemplate(w, indexTmpl, nil)
}

type tokenTmplData struct {
	IDToken      string
	AccessToken  string
	RefreshToken string
	RedirectURL  string
	Claims       string
}

type LoginTmplData struct {
	ReqPath      string
	Username	string
	UsernamePrompt	string
	PostURL	string
	Invalid bool
	Connectors []*ConnectorInfo

}

func getFuncMap() template.FuncMap {
	funcs := sprig.FuncMap()
	additionalFuncs := map[string]interface{}{
		"url": func(reqPath, assetPath string) string {
			return reqPath+ assetPath
		},
	}

	for k, v := range additionalFuncs {
		funcs[k] = v
	}

	return funcs
}

var loginTmplStr =`
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="{{ url .ReqPath "/static/main.css" }}" rel="stylesheet">
    <link href="{{ url .ReqPath "/theme/styles.css" }}" rel="stylesheet">
    <link rel="icon" href="{{ url .ReqPath "/theme/favicon.png" }}">
  </head>

  <body class="theme-body">
    <div class="theme-navbar">
    </div>

    <div class="dex-container">


<div class="theme-panel">
  <h2 class="theme-heading">Log in to Your Account</h2>
  <form method="post" action="{{ .PostURL }}">
    <div class="theme-form-row">
      <div class="theme-form-label">
        <label for="userid">{{ .UsernamePrompt }}</label>
      </div>
	  <input tabindex="1" required id="login" name="login" type="text" class="theme-form-input" placeholder="{{ .UsernamePrompt | lower }}" {{ if .Username }} value="{{ .Username }}" {{ else }} autofocus {{ end }}/>
    </div>
    <div class="theme-form-row">
      <div class="theme-form-label">
        <label for="password">Password</label>
      </div>
	  <input tabindex="2" required id="password" name="password" type="password" class="theme-form-input" placeholder="password" {{ if .Invalid }} autofocus {{ end }}/>
    </div>

    {{ if .Invalid }}
      <div id="login-error" class="dex-error-box">
        Invalid {{ .UsernamePrompt }} and password.
      </div>
    {{ end }}

    <button tabindex="3" id="submit-login" type="submit" class="dex-btn theme-btn--primary">Login</button>

  </form>
  <div >
    {{ range $c := .Connectors }}
        <a href="{{ $c.URL }}"><span class="dex-btn-icon dex-btn-icon--{{ $c.Type }}"></span>
        </a>
    {{ end }}
  </div>
</div>
  {{ if .BackLink }}
  <div class="theme-link-back">
    <a class="dex-subtle-text" href="{{ .BackLink }}">Select another login method.</a>
  </div>
  {{ end }}
</div>

    </div>
  </body>
</html>


`




var tokenTmpl = template.Must(template.New("token.html").Parse(`<html>
  <head>
    <style>
/* make pre wrap */
pre {
 white-space: pre-wrap;       /* css-3 */
 white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
 white-space: -pre-wrap;      /* Opera 4-6 */
 white-space: -o-pre-wrap;    /* Opera 7 */
 word-wrap: break-word;       /* Internet Explorer 5.5+ */
}
    </style>
  </head>
  <body>
    <p> ID Token: <pre><code>{{ .IDToken }}</code></pre></p>
    <p> Access Token: <pre><code>{{ .AccessToken }}</code></pre></p>
    <p> Claims: <pre><code>{{ .Claims }}</code></pre></p>
	{{ if .RefreshToken }}
    <p> Refresh Token: <pre><code>{{ .RefreshToken }}</code></pre></p>
	<form action="{{ .RedirectURL }}" method="post">
	  <input type="hidden" name="refresh_token" value="{{ .RefreshToken }}">
	  <input type="submit" value="Redeem refresh token">
    </form>
	{{ end }}
  </body>
</html>
`))


func renderBiz(w http.ResponseWriter, idToken, accessToken, refreshToken string) {
	renderTemplate(w, bizTmpl, tokenTmplData{
		IDToken:      idToken,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

type Connector struct {
	// ID that will uniquely identify the connector object.
	ID string `json:"id"`
	// The Type of the connector. E.g. 'oidc' or 'ldap'
	Type string `json:"type"`
	// The Name of the connector that is used when displaying it to the end user.
	Name string `json:"name"`
	// ResourceVersion is the static versioning used to keep track of dynamic configuration
	// changes to the connector object made by the API calls.
	ResourceVersion string `json:"resourceVersion"`
}

func renderLogin(w http.ResponseWriter,connectors []*ConnectorInfo,issuerURL string) {
	login,_:= template.New("").Funcs(getFuncMap()).Parse(loginTmplStr)
	renderTemplate(w, login, LoginTmplData{
		ReqPath:  issuerURL,
		UsernamePrompt: "Username",
		PostURL:	"ddd",
		Username:	"",
		Connectors:	connectors,
	})
}

func renderToken(w http.ResponseWriter, redirectURL, idToken, accessToken, refreshToken, claims string) {
	renderTemplate(w, tokenTmpl, tokenTmplData{
		IDToken:      idToken,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		RedirectURL:  redirectURL,
		Claims:       claims,
	})
}

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	err := tmpl.Execute(w, data)
	if err == nil {
		return
	}

	switch err := err.(type) {
	case *template.Error:
		// An ExecError guarantees that Execute has not written to the underlying reader.
		log.Printf("Error rendering template %s: %s", tmpl.Name(), err)

		// TODO(ericchiang): replace with better internal server error.
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	default:
		// An error with the underlying write, such as the connection being
		// dropped. Ignore for now.
	}
}
