package userpass

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/policyutil"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathUserPolicies(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "users/" + framework.GenericNameRegex("username") + "/policies$",
		Fields: map[string]*framework.FieldSchema{
			"username": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Username for this user.",
			},
			"policies": &framework.FieldSchema{
				Type:        framework.TypeCommaStringSlice,
				Description: "(DEPRECATED) Use 'token_policies' instead. If this and 'token_policies' are both specified only 'token_policies' will be used.",
				Deprecated:  true,
			},
			"token_policies": &framework.FieldSchema{
				Type:        framework.TypeCommaStringSlice,
				Description: "Comma-separated list of policies",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathUserPoliciesUpdate,
		},

		HelpSynopsis:    pathUserPoliciesHelpSyn,
		HelpDescription: pathUserPoliciesHelpDesc,
	}
}

func (b *backend) pathUserPoliciesUpdate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	username := d.Get("username").(string)

	userEntry, err := b.user(ctx, req.Storage, username)
	if err != nil {
		return nil, err
	}
	if userEntry == nil {
		return nil, fmt.Errorf("username does not exist")
	}

	var resp *logical.Response

	policiesRaw, ok := d.GetOk("token_policies")
	if !ok {
		policiesRaw, ok = d.GetOk("policies")
		if ok {
			userEntry.Policies = policyutil.ParsePolicies(policiesRaw)
			userEntry.TokenPolicies = nil
		}
	} else {
		userEntry.TokenPolicies = policyutil.ParsePolicies(policiesRaw)
		_, ok = d.GetOk("policies")
		if ok {
			if resp == nil {
				resp = &logical.Response{}
			}
			resp.AddWarning("Both 'token_policies' and deprecated 'policies' values supplied, ignoring the deprecated value")
		}
		userEntry.Policies = nil
	}

	return resp, b.setUser(ctx, req.Storage, username, userEntry)
}

const pathUserPoliciesHelpSyn = `
Update the policies associated with the username.
`

const pathUserPoliciesHelpDesc = `
This endpoint allows updating the policies associated with the username.
`
