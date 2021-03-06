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

package schema

import (
	"context"
	"fmt"

	"github.com/semi-technologies/weaviate/entities/models"
	"github.com/semi-technologies/weaviate/entities/schema/kind"
)

// DeleteAction Class to the schema
func (m *Manager) DeleteAction(ctx context.Context, principal *models.Principal, class string) error {
	err := m.authorizer.Authorize(principal, "delete", "schema/actions")
	if err != nil {
		return err
	}

	return m.deleteClass(ctx, class, kind.Action)
}

// DeleteThing Class to the schema
func (m *Manager) DeleteThing(ctx context.Context, principal *models.Principal, class string) error {
	err := m.authorizer.Authorize(principal, "delete", "schema/things")
	if err != nil {
		return err
	}

	return m.deleteClass(ctx, class, kind.Thing)
}

func (m *Manager) deleteClass(ctx context.Context, className string, k kind.Kind) error {
	unlock, err := m.locks.LockSchema()
	if err != nil {
		return err
	}
	defer unlock()

	semanticSchema := m.state.SchemaFor(k)
	var classIdx = -1
	for idx, class := range semanticSchema.Classes {
		if class.Class == className {
			classIdx = idx
			break
		}
	}

	if classIdx == -1 {
		return fmt.Errorf("could not find class '%s'", className)
	}

	semanticSchema.Classes[classIdx] = semanticSchema.Classes[len(semanticSchema.Classes)-1]
	semanticSchema.Classes[len(semanticSchema.Classes)-1] = nil // to prevent leaking this pointer.
	semanticSchema.Classes = semanticSchema.Classes[:len(semanticSchema.Classes)-1]

	err = m.saveSchema(ctx)
	if err != nil {
		return err
	}

	return m.migrator.DropClass(ctx, k, className)
	//TODO gh-846: rollback state update if migration fails
}
