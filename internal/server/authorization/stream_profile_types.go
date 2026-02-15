package authorization

import (
	"strings"
)

// Internal Profile struct, do not use for endpoints
type Profile struct {
	FileName string
	IsActive bool
	IsPublic bool
	MOTD     string
}

var separator = "_"

func (profile *Profile) StreamKey() string {
	splitIndex := strings.LastIndex(profile.FileName, separator)
	return profile.FileName[:splitIndex+len(separator)-1]
}
func (profile *Profile) StreamToken() string {
	splitIndex := strings.LastIndex(profile.FileName, separator)
	return profile.FileName[splitIndex+len(separator):]
}
func (profile *Profile) AsPublicProfile() *PublicProfile {
	return &PublicProfile{
		StreamKey: profile.StreamKey(),
		IsActive:  profile.IsActive,
		IsPublic:  profile.IsPublic,
		MOTD:      profile.MOTD,
	}
}
func (profile *Profile) AsPersonalProfile() *PersonalProfile {
	return &PersonalProfile{
		StreamKey: profile.StreamKey(),
		IsActive:  profile.IsActive,
		IsPublic:  profile.IsPublic,
		MOTD:      profile.MOTD,
	}
}
func (profile *Profile) AsAdminProfile() *AdminProfile {
	return &AdminProfile{
		StreamKey: profile.StreamKey(),
		Token:     profile.StreamToken(),
		IsPublic:  profile.IsPublic,
		MOTD:      profile.MOTD,
	}
}

// Public profile struct for serving to public endpoints
type PublicProfile struct {
	StreamKey string `json:"streamKey"`
	IsActive  bool   `json:"isActive"`
	IsPublic  bool   `json:"isPublic"`
	MOTD      string `json:"motd"`
}

// Personal profile struct for serving to profile owner endpoints
type PersonalProfile struct {
	StreamKey string `json:"streamKey"`
	IsActive  bool   `json:"isActive"`
	IsPublic  bool   `json:"isPublic"`
	MOTD      string `json:"motd"`
}

// Admin profile struct for serving to admin specific endpoints
type AdminProfile struct {
	StreamKey string `json:"streamKey"`
	Token     string `json:"token"`
	IsPublic  bool   `json:"isPublic"`
	MOTD      string `json:"motd"`
}
