/*
 * Copyright 2019 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
 * Use of this source code is governed by an MIT-style license that can be found in the LICENSE file
 */

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gitlab.myteksi.net/product-security/ssdlc/secret-scanner/web"

	"gitlab.myteksi.net/product-security/ssdlc/secret-scanner/scanner"
	"gitlab.myteksi.net/product-security/ssdlc/secret-scanner/scanner/gitprovider"
	"gitlab.myteksi.net/product-security/ssdlc/secret-scanner/scanner/options"
	"gitlab.myteksi.net/product-security/ssdlc/secret-scanner/scanner/session"
)

func main() {
	opt, err := options.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Validate Options
	optValid, err := opt.ValidateOptions()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if !optValid {
		fmt.Println(errors.New("invalid option(s)"))
		os.Exit(1)
	}

	var gitProvider gitprovider.GitProvider
	additionalParams := map[string]string{}

	// Set Git provider
	switch *opt.GitProvider {
	case gitprovider.GithubName:
		gitProvider = &gitprovider.GithubProvider{}
	case gitprovider.GitlabName:
		gitProvider = &gitprovider.GitlabProvider{}
	case gitprovider.BitbucketName:
		gitProvider = &gitprovider.BitbucketProvider{}
	default:
		fmt.Println("error: invalid Git provider type (Currently supports github, gitlab, bitbucket)")
		os.Exit(1)
	}

	// Initialize Git provider
	err = gitProvider.Initialize(*opt.BaseURL, *opt.Token, additionalParams)
	if err != nil {
		fmt.Println(errors.New(fmt.Sprintf("unable to initialise %s provider", *opt.GitProvider)))
		os.Exit(1)
	}

	// Initialize new scan session
	sess := &session.Session{}
	sess.Initialize(opt)
	sess.Out.Important("%s Scanning Started at %s\n", strings.Title(*opt.GitProvider), sess.Stats.StartedAt.Format(time.RFC3339))
	sess.Out.Important("Loaded %d signatures\n", len(sess.Signatures))

	if sess.Stats.Status == "finished" {
		sess.Out.Important("Loaded session file: %s\n", *sess.Options.Load)
		return
	}

	// Scan
	scanner.Scan(sess, gitProvider)
	sess.Out.Important("Gitlab Scanning Finished at %s\n", sess.Stats.FinishedAt.Format(time.RFC3339))

	if *sess.Options.Report != "" {
		err := sess.SaveToFile(*sess.Options.Report)
		if err != nil {
			sess.Out.Error("Error saving session to %s: %s\n", *sess.Options.Report, err)
		}
		sess.Out.Important("Saved session to: %s\n\n", *sess.Options.Report)
	}

	sess.Stats.PrintStats(sess.Out)

	// Serve UI
	if *sess.Options.UI {
		web.InitRouter("127.0.0.1", "8888", sess)
	}
}
