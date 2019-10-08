package writers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/jtarchie/dothings/status"

	"github.com/buildkite/terminal-to-html"
	"github.com/jtarchie/dothings/executor"
	"github.com/jtarchie/dothings/planner"
)

type handler struct {
	plan   planner.Step
	writer executor.Writer
	stater status.Stater
}

func NewWebHandler(
	plan planner.Step,
	writer executor.Writer,
	stater status.Stater,
) *handler {
	return &handler{
		plan:   plan,
		writer: writer,
		stater: stater,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	currentStatus := h.plan.State(h.stater)
	_, _ = fmt.Fprintf(w, `<html>
	<head>
		<meta charset="utf-8">
		<meta name="turbolinks-cache-control" content="no-preview">
		<script src="https://cdnjs.cloudflare.com/ajax/libs/turbolinks/5.2.0/turbolinks.js" integrity="sha256-iM4Yzi/zLj/IshPWMC1IluRxTtRjMqjPGd97TZ9yYpU=" crossorigin="anonymous"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/tinycon/0.6.5/tinycon.min.js" integrity="sha256-mWm+F2y8RzP5KPoO6HueiXPFJwPj8nIuwICaExSTU10=" crossorigin="anonymous"></script>
		<link rel="stylesheet" href="https://rawcdn.githack.com/buildkite/terminal-to-html/697ff23bd8dc48b9d23f11f259f5256dae2455f0/assets/terminal.css">
		<link rel="stylesheet" href="https://unpkg.com/picnic">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<style>
			.container { padding: 20px }
			.type-serial {
				padding: 5px;
				background-color: rgba(0,31,63,0.2);
			}
			.type-parallel {
				padding: 5px;
				background-color: rgba(1,255,112, 0.2);
			}
			.type-task {
				background-color: #fff;
			}
			.status .id:before {
				margin-right: 10px;
				height: 17px;
				width: 17px;
				border-radius: 50%%;
				display: inline-block;
				content: ' ';
				background-color: gray;
			}
			.status.success .id:before {
				background-color: #2ecc40;
			}
			.status.errored .id:before {
				background-color: #f5a623; 
			}
			.status.unstarted .id:before {
				background-color: #bbb;
			}
			.status.running .id:before {
				background-color: #0074d9;
			}
			.status.failed .id:before {
				background-color: #ff4136;
			}
		</style>
	</head>
	<body>
	<div class="container %s">`, currentStatus)
	h.writeTree(w, h.plan.Tree())
	_, _ = fmt.Fprintf(w, `
	</div>
	<script>
		(function(){
			var reloadWithTurbolinks = (function () {
			  var scrollPosition
			
			  function reload () {
				scrollPosition = [window.scrollX, window.scrollY]
				Turbolinks.visit(window.location.toString(), { action: 'replace' })
			  }
			
			  document.addEventListener('turbolinks:load', function () {
				if (scrollPosition) {
				  window.scrollTo.apply(window, scrollPosition)
				  scrollPosition = null
				}
			  })
			
			  return reload
			})()
            var status = "%s";
			if (status == "unstarted" || status == "running") {
				console.log("refreshing view");
				setTimeout(reloadWithTurbolinks, 500);
			}
		})();
	</script>
	</body></html>
	`, currentStatus)
}

func (h *handler) writeTree(
	writer io.Writer,
	tree planner.Tree,
) {

	if tree.Type() == planner.Task {
		status := h.stater.Get(tree.Task())
		if len(status) > 0 {
			_, _ = fmt.Fprintf(writer, `<article class="card type-%s status %s">`, tree.Type(), status[0])
		} else {
			_, _ = fmt.Fprintf(writer, `<article class="card type-%s status">`, tree.Type())
		}

		stdout, _ := h.writer.GetString(tree.Task())
		_, _ = fmt.Fprintf(writer, `<header class="id">%s</header>`, tree.Task().ID())
		_, _ = fmt.Fprintf(
			writer,
			`<div class="term-container">%s</div>`,
			terminal.Render(
				[]byte(
					stdout,
				),
			),
		)
	} else {
		_, _ = fmt.Fprintf(writer, `<div class="type-%s">`, tree.Type())
	}
	for _, node := range tree.Children() {
		h.writeTree(writer, node)
	}
	_, _ = fmt.Fprint(writer, "</article>")
}
