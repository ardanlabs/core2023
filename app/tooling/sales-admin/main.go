package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/open-policy-agent/opa/rego"
)

func main() {
	err := GenToken()
	if err != nil {
		log.Fatalln(err)
	}
}

// GenToken generates a JWT for the specified user.
func GenToken() error {

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := struct {
		jwt.RegisteredClaims
		Roles []string
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "123456789",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{"USER"},
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS256.Name)

	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	file, err := os.Open("zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem")
	if err != nil {
		return fmt.Errorf("opening key file: %w", err)
	}
	defer file.Close()

	// limit PEM file size to 1 megabyte. This should be reasonable for
	// almost any PEM file and prevents shenanigans like linking the file
	// to /dev/random or something like that.
	pemData, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return fmt.Errorf("reading auth private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemData)
	if err != nil {
		return fmt.Errorf("parsing auth private key: %w", err)
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	fmt.Println(str)
	fmt.Println("=======================================================")

	// -------------------------------------------------------------------------
	// Generating Public Key

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	var b bytes.Buffer
	if err := pem.Encode(&b, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	fmt.Println(b.String())
	fmt.Println("=======================================================")

	// -------------------------------------------------------------------------
	// Decode Token To Get Parts

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))

	var claims2 struct {
		jwt.RegisteredClaims
		Roles []string
	}
	_, _, err = parser.ParseUnverified(str, &claims2)
	if err != nil {
		return fmt.Errorf("error parsing token: %w", err)
	}

	fmt.Printf("%#v\n", claims2)
	fmt.Println("=======================================================")

	// -------------------------------------------------------------------------
	// Authentication in Rego

	input := map[string]any{
		"Key":   b.String(),
		"Token": str,
		"ISS":   "service project",
	}

	query := fmt.Sprintf("x = data.%s.%s", "ardan.rego", "auth")

	ctx := context.Background()

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", authentication),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		fmt.Println("Signature Not Validated")
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	fmt.Println("Validated Signature")
	fmt.Println("=======================================================")

	// -------------------------------------------------------------------------
	// Authorization in Rego

	input2 := map[string]any{
		"Roles": []string{"ADMIN"},
	}

	query2 := fmt.Sprintf("x = data.%s.%s", "ardan.rego", "rule_admin_only")

	q2, err := rego.New(
		rego.Query(query2),
		rego.Module("policy.rego", authorization),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	results2, err := q2.Eval(ctx, rego.EvalInput(input2))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results2) == 0 {
		return errors.New("no results")
	}

	result2, ok := results2[0].Bindings["x"].(bool)
	if !ok || !result2 {
		fmt.Println("User NOT ADIMN - NOT AUTHORIZED")
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	fmt.Println("AUTHORIZED")

	return nil
}

var authorization = `package ardan.rego

import future.keywords.if
import future.keywords.in

default rule_admin_only := false

role_admin := "ADMIN"

rule_admin_only if {
	claim_roles := {role | some role in input.Roles}
	input_admin := {role_admin} & claim_roles
	count(input_admin) > 0
}`

var authentication = `package ardan.rego

import future.keywords.if
import future.keywords.in

default auth := false

auth if {
	[valid, _, _] := verify_jwt
	valid = true
}

verify_jwt := io.jwt.decode_verify(input.Token, {
	"cert": input.Key,
	"iss": input.ISS,
})`

// GenKey creates an x509 private/public key for auth tokens.
func GenKey() error {

	// Generate a new private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	// Create a file for the private key information in PEM form.
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return fmt.Errorf("creating private file: %w", err)
	}
	defer privateFile.Close()

	// Construct a PEM block for the private key.
	privateBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write the private key to the private key file.
	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}

	// -------------------------------------------------------------------------

	// Create a file for the public key information in PEM form.
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	fmt.Println("private and public key files generated")

	return nil
}
