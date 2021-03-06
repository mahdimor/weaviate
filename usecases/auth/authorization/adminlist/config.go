//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2019 SeMI Holding B.V. (registered @ Dutch Chamber of Commerce no 75221632). All rights reserved.
//  LICENSE WEAVIATE OPEN SOURCE: https://www.semi.technology/playbook/playbook/contract-weaviate-OSS.html
//  LICENSE WEAVIATE ENTERPRISE: https://www.semi.technology/playbook/contract-weaviate-enterprise.html
//  CONCEPT: Bob van Luijt (@bobvanluijt)
//  CONTACT: hello@semi.technology
//

package adminlist

import "fmt"

// Config makes every subject on the list an admin, whereas everyone else
// has no rights whatsoever
type Config struct {
	Enabled       bool     `json:"enabled" yaml:"enabled"`
	Users         []string `json:"users" yaml:"users"`
	ReadOnlyUsers []string `json:"read_only_users" yaml:"read_only_users"`
}

// Validate admin list config for viability, can be called from the central
// config package
func (c Config) Validate() error {
	return c.validateOverlap()
}

// we are expecting both lists to always contain few subjects and know that
// this comparison is only done once (at startup). We are therefore fine with
// the O(n^2) complexity of this very primitive overlap search in favor of very
// simple code.
func (c Config) validateOverlap() error {
	for _, a := range c.Users {
		for _, b := range c.ReadOnlyUsers {
			if a == b {
				return fmt.Errorf("admin list: subject '%s' is present on both admin and read-only list", a)
			}
		}
	}

	return nil
}
