swagger: 
	swag generate spec -o ./swagger.json
	swag serve -F=swagger swagger.json