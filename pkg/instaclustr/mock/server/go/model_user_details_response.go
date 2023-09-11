/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// UserDetailsResponse - The details of the current user.
type UserDetailsResponse struct {

	// The unique identifier for this user.
	Id string `json:"id,omitempty"`

	// The user's preferred name for communication.
	DisplayName string `json:"displayName,omitempty"`

	// The user's email address.
	Email string `json:"email,omitempty"`

	IsAdmin bool `json:"isAdmin,omitempty"`
}

// AssertUserDetailsResponseRequired checks if the required fields are not zero-ed
func AssertUserDetailsResponseRequired(obj UserDetailsResponse) error {
	return nil
}

// AssertUserDetailsResponseConstraints checks if the values respects the defined constraints
func AssertUserDetailsResponseConstraints(obj UserDetailsResponse) error {
	return nil
}
