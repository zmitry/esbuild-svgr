package svgr

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/JoshVarga/svgparser"
	"github.com/evanw/esbuild/pkg/api"
)

func svgToJsx(items []*svgparser.Element) string {
	res := ""
	for _, el := range items {
		str := fmt.Sprintf("<%s %s>%s</%s>", el.Name, attributesToString(el.Attributes), svgToJsx(el.Children), el.Name)
		res += str
	}
	return res
}

func attributesToString(attr map[string]string) string {
	res := ""
	for key, val := range attr {
		str := fmt.Sprintf(`%s="%s"`, key, val)
		res += str
	}
	return res
}

// Plugin can be used in api.BuildOptions.
func SVGRPlugin(plugin api.Plugin) {
	plugin.SetName("SVGR")
	plugin.AddResolver(api.ResolverOptions{Filter: "^svgr:"}, func(args api.ResolverArgs) (res api.ResolverResult, err error) {
		res.Path = path.Join(path.Dir(args.Importer), strings.TrimLeft(args.Path, "svgr:"))
		res.Namespace = "svg"
		return res, nil
	})
	plugin.AddLoader(
		api.LoaderOptions{Filter: `\.svg`, Namespace: "svg"},
		func(args api.LoaderArgs) (res api.LoaderResult, err error) {
			dat, err := os.Open(args.Path)
			if err != nil {
				return res, err
			}

			svg, err := svgparser.Parse(dat, true)
			if err != nil {
				return res, err
			}
			contents := fmt.Sprintf(`
      import React from "react";
      import url from "%s"
      export default url;
      export function ReactComponent(props) {
        return <svg %s {...props}>
        %s
        </svg>
      }
      `, args.Path, attributesToString(svg.Attributes), svgToJsx(svg.Children))
			res.Contents = &contents
			res.Loader = api.LoaderTSX
			res.ResolveDir = path.Dir(args.Path)
			return res, nil
		},
	)
}
