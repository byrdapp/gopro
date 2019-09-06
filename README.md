

## Old claims for JWT (broken)

	// fbServiceAcc := os.Getenv("FB_SERVICE_ACC_EMAIL")
		// token := jwt.New(jwt.GetSigningMethod(jwt.SigningMethodRS256.Alg()))
		// mapclaims := token.Claims.(jwt.MapClaims)
		// mapclaims["alg"] = "RS256"
		// mapclaims["iss"] = fbServiceAcc
		// mapclaims["sub"] = fbServiceAcc
		// mapclaims["aud"] = "https://identitytoolkit.googleapis.com/google.identity.identitytoolkit.v1.IdentityToolkit"
		// mapclaims["iat"] = time.Now().Unix()
		// mapclaims["exp"] = time.Now().Add(time.Minute * 30).Unix()
		// mapclaims["uid"] = usr.UID

		// signedToken, err := token.SignedString(JWTKeyPrivate())
		// if err != nil {
		// 	errors.NewResErr(err, "Could not sign token", http.StatusInternalServerError, w)
		// 	return
		// }

// claims := &Claims{
		// 	JWTClaims: jwt.StandardClaims{
		// 		IssuedAt:  time.Now().Unix(),
		// 		ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
		// 		Audience:  "https://identitytoolkit.googleapis.com/google.identity.identitytoolkit.v1.IdentityToolkit",
		// 		Subject:   os.Getenv("FB_SERVICE_ACC_EMAIL"),
		// 		Issuer:    os.Getenv("FB_SERVICE_ACC_EMAIL"),
		// 	},
		// 	UID: creds.UID,
		// }
		// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims.JWTClaims)


		<!-- token, err := jwt.ParseWithClaims(cookie.Value, &claims.JWTClaims, func(token *jwt.Token) (interface{}, error) {
			fmt.Println(cookie.Value)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				errors.NewResErr(err, err.Error(), http.StatusInternalServerError, w)
				return nil, err
			}
			return JWTKeyPublic(), nil
		}) -->


### Considering using PostForm for this:
#### Create booking
	// if err := r.ParseForm(); err != nil {
		// 	errors.NewResErr(err, err.Error(), 503, w)
		// 	return
		// }

		// req.Lat = r.PostFormValue("lat")
		// req.Lng = r.PostFormValue("lng")
		// req.Task = r.PostFormValue("task")
		// req.Booker = r.PostFormValue("booker")
		// req.Price = r.PostFormValue("price")
		// req.Credits = r.PostFormValue("credits")
		// req.DateEnd = r.PostFormValue("dateEnd")
		// req.DateStart = r.PostFormValue("dateStart")

		// for key := range r.Form {
		// 	str := r.PostFormValue(key)
		// 	log.Infof("Val %s", str)
		// }