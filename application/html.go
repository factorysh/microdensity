package application

import "io"

func writeHTMLHeader(w io.Writer) {
	w.Write([]byte(`
	<!doctype html>
	<html>
	  <head>
	    <meta charset="UTF-8">
	    <meta name="viewport" content="width=device-width, initial-scale=1.0">
	    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.3/css/bulma.min.css">
	    <title>µdensity</title>
	  </head>
	  <body>
	    <section class="section">
	      <div class="container">
	`))
}

func writeHTMLFooter(w io.Writer) {
	w.Write([]byte(`
	      </div>
	    </section>
	  </body>
	</html>
	`))
}
