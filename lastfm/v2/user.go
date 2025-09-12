package lfm

type userApi struct {
	api *LastFMApi
}

// user.getInfo
func (u *userApi) GetInfo(args P) (*UserGetInfo, error) {
	req := u.api.baseRequest("user.getinfo", args)

	data, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var result UserGetInfo
	if err := decodeResponse(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
