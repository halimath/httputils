package securityheader

import (
	"net/http"
	"strings"
)

const (
	// CSP directive value none
	CSPNone = "'none'"
	// CSP directive value self
	CSPSelf = "'self'"
	// CSP directive value unsafe-inline
	CSPUnsafeInline = "'unsafe-inline'"
	// CSP directive value unsafe-eval
	CSPUnsafeEval = "'unsafe-eval'"
	// CSP directive value wasm-unsafe-eval
	CSPWASMUnsafeEval = "'wasm-unsafe-eval'"
	// CSP directive value trusted-types-eval
	CSPTrustedTypesEval = "'trusted-types-eval'"
	// CSP directive value inline-speculation-rules
	CSPInlineSpeculationRules = "'inline-speculation-rules'"
	// CSP directive value strict-dynamic
	CSPStrictDynamic = "'strict-dynamic'"
)

type cspDirective string

const (
	// Defines the valid sources for web workers and nested browsing contexts loaded using elements such as <frame> and <iframe>.
	// Fallback for frame-src and worker-src.
	CSPChildSrc cspDirective = "child-src"

	// Restricts the URLs which can be loaded using script interfaces.
	CSPConnectSrc cspDirective = "connect-src"

	// Serves as a fallback for the other fetch directives.
	// Fallback for all other fetch directives.
	CSPDefaultSrc cspDirective = "default-src"

	// Specifies valid sources for nested browsing contexts loaded into <fencedframe> elements.
	CSPFencedFrameSrc cspDirective = "fenced-frame-src"

	// Specifies valid sources for fonts loaded using @font-face.
	CSPFontSrc cspDirective = "font-src"

	// Specifies valid sources for nested browsing contexts loaded into elements such as <frame> and <iframe>.
	CSPFrameSrc cspDirective = "frame-src"

	// Specifies valid sources of images and favicons.
	CSPImgSrc cspDirective = "img-src"

	// Specifies valid sources of application manifest files.
	CSPManifestSrc cspDirective = "manifest-src"

	// Specifies valid sources for loading media using the <audio>, <video> and <track> elements.
	CSPMediaSrc cspDirective = "media-src"

	// Specifies valid sources for the <object> and <embed> elements.
	CSPObjectSrc cspDirective = "object-src"

	// Specifies valid sources for JavaScript and WebAssembly resources.
	// Fallback for script-src-elem and script-src-attr.
	CSPScriptSrc cspDirective = "script-src"

	// Specifies valid sources for JavaScript <script> elements.
	CSPScriptSrcElem cspDirective = "script-src-elem"

	// Specifies valid sources for JavaScript inline event handlers.
	CSPScriptSrcAttr cspDirective = "script-src-attr"

	// Specifies valid sources for stylesheets.
	// Fallback for style-src-elem and style-src-attr.
	CSPStyleSrc cspDirective = "style-src"

	// Specifies valid sources for stylesheets <style> elements and <link> elements with rel="stylesheet".
	CSPStyleSrcElem cspDirective = "style-src-elem"

	// Specifies valid sources for inline styles applied to individual DOM elements.
	CSPStyleSrcAttr cspDirective = "style-src-attr"

	// Specifies valid sources for Worker, SharedWorker, or ServiceWorker scripts.
	CSPWorkerSrc cspDirective = "worker-src"
)

type cspPolicyDirective struct {
	directive cspDirective
	values    []string
}

func (cpd cspPolicyDirective) headerValue() string {
	return string(cpd.directive) + " " + strings.Join(cpd.values, " ")
}

// Factory for a single CSP policy directive defining values as valid sources for directive.
func CSPPolicyDirective(directive cspDirective, values ...string) cspPolicyDirective {
	return cspPolicyDirective{directive, values}
}

// Configures a middleware Option to set the Content-Security-Policy header
// based on the given policyDirectives. If policyDirectives is empty,
//
//	default-src 'self'
//
// is used.
func ContentSecurityPolicy(policyDirectives ...cspPolicyDirective) Option {
	if len(policyDirectives) == 0 {
		policyDirectives = append(policyDirectives, CSPPolicyDirective(CSPDefaultSrc, CSPSelf))
	}

	headerValue := joinDirectives(policyDirectives)

	return func(h http.Header) {
		h.Set("content-security-policy", headerValue)
	}
}
