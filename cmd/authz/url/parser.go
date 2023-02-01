package url

import (
	"fmt"
	"github.com/myeong01/ai-playground/pkg/authz/authorizer"
	"net/http"
	"strings"
)

func UrlToChecker(url, method string) (*authorizer.Checker, error) {
	clearUrl := strings.TrimPrefix(url, "/api/resource/")
	splitUrl := removeZeroValue(strings.Split(clearUrl, "/"))
	return urlToChecker(url, splitUrl, 0, method)
}

func removeZeroValue(strs []string) []string {
	newStrs := make([]string, 0)
	for _, str := range strs {
		if str != "" {
			newStrs = append(newStrs, str)
		}
	}
	return newStrs
}

func urlToChecker(url string, splitUrl []string, curIndex int, method string) (*authorizer.Checker, error) {
	if len(splitUrl) <= curIndex+2 {
		return nil, ErrBadRequest.WithMsg(fmt.Sprintf("invalid url path: [%s]", url))
	}
	wide := splitUrl[curIndex+2]
	var nextFn func(url string, splitUrl []string, curIndex int, method string) (*authorizer.Checker, error)
	switch wide {
	case "cluster":
		nextFn = urlToCheckerInResource
	case "namespace":
		nextFn = urlToCheckerInNamespace
	default:
		return nil, ErrBadRequest.WithMsg(fmt.Sprintf("invalid url path: [%s], path next to version must be \"cluster\" or \"namespace\", but got [%s]", url, wide))
	}
	checker, err := nextFn(url, splitUrl, curIndex+3, method)
	if err != nil {
		return nil, err
	}
	return checker.
		WithGroup(splitUrl[curIndex]).
		WithVersion(splitUrl[curIndex+1]), nil
}

func urlToCheckerInResource(url string, splitUrl []string, curIndex int, method string) (*authorizer.Checker, error) {
	splitUrlLen := len(splitUrl)
	if splitUrlLen <= curIndex {
		return nil, ErrBadRequest.WithMsg(fmt.Sprintf("invalid url path: [%s]", url))
	}
	if splitUrlLen == curIndex+1 {
		return addVerbByMethod(authorizer.NewChecker().WithResource(splitUrl[curIndex]), url, method, false)
	}

	wide := splitUrl[curIndex+1]
	var nextFn func(url string, splitUrl []string, curIndex int, method string) (*authorizer.Checker, error)
	switch splitUrl[curIndex+1] {
	case "object":
		nextFn = urlToCheckerInObject
	case "subresource":
		nextFn = urlToCheckerInSubresource
	default:
		return nil, ErrBadRequest.WithMsg(fmt.Sprintf("invalid url path: [%s], path next to resource must be \"subresource\" or \"object\", but got [%s]", url, wide))
	}
	checker, err := nextFn(url, splitUrl, curIndex+2, method)
	if err != nil {
		return nil, err
	}
	return checker.WithResource(splitUrl[curIndex]), nil
}

func urlToCheckerInNamespace(url string, splitUrl []string, curIndex int, method string) (*authorizer.Checker, error) {
	splitUrlLen := len(splitUrl)
	if splitUrlLen <= curIndex+1 {
		return nil, ErrBadRequest.WithMsg(fmt.Sprintf("invalid url path: [%s]", url))
	}
	wide := splitUrl[curIndex+1]
	var nextFn func(url string, splitUrl []string, curIndex int, method string) (*authorizer.Checker, error)
	switch wide {
	case "group":
		nextFn = urlToCheckerInGroup
	case "playground":
		nextFn = urlToCheckerInResource
	default:
		return nil, ErrBadRequest.WithMsg(fmt.Sprintf("invalid url path: [%s], path next to version must be \"cluster\" or \"namespace\", but got [%s]", url, wide))
	}
	checker, err := nextFn(url, splitUrl, curIndex+2, method)
	if err != nil {
		return nil, err
	}
	return checker.WithObjectNamespace(splitUrl[curIndex]), nil
}

func urlToCheckerInObject(url string, splitUrl []string, curIndex int, method string) (*authorizer.Checker, error) {
	splitUrlLen := len(splitUrl)
	if splitUrlLen <= curIndex {
		return nil, ErrBadRequest.WithMsg(fmt.Sprintf("invalid url path: [%s]", url))
	}
	return addVerbByMethod(authorizer.NewChecker().WithObjectName(splitUrl[curIndex]), url, method, true)
}

func urlToCheckerInSubresource(url string, splitUrl []string, curIndex int, method string) (*authorizer.Checker, error) {
	splitUrlLen := len(splitUrl)
	if splitUrlLen <= curIndex {
		return nil, ErrBadRequest.WithMsg(fmt.Sprintf("invalid url path: [%s]", url))
	}
	if splitUrlLen == curIndex+1 {
		return addVerbByMethod(authorizer.NewChecker().WithSubresource(splitUrl[curIndex]), url, method, false)
	}
	checker, err := urlToCheckerInObject(url, splitUrl, curIndex+1, method)
	if err != nil {
		return nil, err
	}
	return checker.WithSubresource(splitUrl[curIndex]), nil
}

func urlToCheckerInGroup(url string, splitUrl []string, curIndex int, method string) (*authorizer.Checker, error) {
	splitUrlLen := len(splitUrl)
	if splitUrlLen <= curIndex+1 {
		return nil, ErrBadRequest.WithMsg(fmt.Sprintf("invalid url path: [%s]", url))
	}

	checker, err := urlToCheckerInResource(url, splitUrl, curIndex+1, method)
	if err != nil {
		return nil, err
	}
	return checker.WithGroupWide(splitUrl[curIndex]), nil
}

func addVerbByMethod(checker *authorizer.Checker, url, method string, isEndWithObjectName bool) (*authorizer.Checker, error) {
	if isEndWithObjectName {
		switch method {
		case http.MethodGet:
			return checker.WithVerb("get"), nil
		case http.MethodDelete:
			return checker.WithVerb("delete"), nil
		case http.MethodPut:
			return checker.WithVerb("update"), nil
		}
	} else {
		switch method {
		case http.MethodGet:
			return checker.WithVerb("list"), nil
		case http.MethodPost:
			return checker.WithVerb("create"), nil
		}
	}
	return nil, ErrBadRequest.WithMsg(fmt.Sprintf("unsupported method [%s] for url [%s]", method, url))
}

// GET 		/				- list
// GET 		/<objectName>	- get
// POST		/				- create
// DELETE	/<objectName>	- delete
// PUT		/<objectName>	- update

// /api/resource/<group>/<version>/namespace/<namespace>/group/<group>/<resource>/subresource/<subresource>/<objectName>	<= 10
// /api/resource/<group>/<version>/namespace/<namespace>/group/<group>/<resource>/subresource/<subresource>					<= 9
// /api/resource/<group>/<version>/namespace/<namespace>/group/<group>/<resource>/object/<objectName>						<= 9
// /api/resource/<group>/<version>/namespace/<namespace>/group/<group>/<resource>											<= 7
// /api/resource/<group>/<version>/namespace/<namespace>/playground/<resource>/subresource/<subresource>/<objectName>		<= 9
// /api/resource/<group>/<version>/namespace/<namespace>/playground/<resource>/subresource/<subresource>					<= 8
// /api/resource/<group>/<version>/namespace/<namespace>/playground/<resource>/object/<objectName>							<= 8
// /api/resource/<group>/<version>/namespace/<namespace>/playground/<resource>												<= 6
// /api/resource/<group>/<version>/cluster/<resource>/subresource/<subresource>/<objectName> 								<= 7
// /api/resource/<group>/<version>/cluster/<resource>/subresource/<subresource> 											<= 6
// /api/resource/<group>/<version>/cluster/<resource>/object/<objectName> 													<= 6
// /api/resource/<group>/<version>/cluster/<resource> 																		<= 4

// <namespace>-<group name>-<role>
