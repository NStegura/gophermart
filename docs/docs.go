// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/user/balance": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "get user balance",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get balance",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_NStegura_gophermart_internal_app_gophermartapi_models.Balance"
                        }
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/api/user/balance/withdraw": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "create user withdraw",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Create withdraw",
                "parameters": [
                    {
                        "description": "User withdraw data",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_NStegura_gophermart_internal_app_gophermartapi_models.WithdrawIn"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "402": {
                        "description": "Payment Required"
                    },
                    "422": {
                        "description": "Unprocessable Entity"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/api/user/login": {
            "post": {
                "description": "login",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Login",
                "parameters": [
                    {
                        "description": "User data",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_NStegura_gophermart_internal_app_gophermartapi_models.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "headers": {
                            "Authorization": {
                                "type": "string",
                                "description": "Use this header in other endpoints"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/api/user/orders": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "get order list by user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get order list",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/github_com_NStegura_gophermart_internal_app_gophermartapi_models.Order"
                            }
                        }
                    },
                    "204": {
                        "description": "No Content"
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "\"register\" user order",
                "consumes": [
                    "text/plain"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Create order",
                "parameters": [
                    {
                        "description": "Order id",
                        "name": "string",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "202": {
                        "description": "Accepted"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "409": {
                        "description": "Conflict"
                    },
                    "422": {
                        "description": "Unprocessable Entity"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/api/user/register": {
            "post": {
                "description": "register",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Register",
                "parameters": [
                    {
                        "description": "User data",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_NStegura_gophermart_internal_app_gophermartapi_models.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "headers": {
                            "Authorization": {
                                "type": "string",
                                "description": "Use this header in other endpoints"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "409": {
                        "description": "Conflict"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/api/user/withdrawals": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "get user withdraw list",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get withdraw list",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/github_com_NStegura_gophermart_internal_app_gophermartapi_models.WithdrawOut"
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "check service",
                "tags": [
                    "tech"
                ],
                "summary": "Get ping",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        }
    },
    "definitions": {
        "github_com_NStegura_gophermart_internal_app_gophermartapi_models.Balance": {
            "type": "object",
            "properties": {
                "current": {
                    "type": "number"
                },
                "withdrawn": {
                    "type": "number"
                }
            }
        },
        "github_com_NStegura_gophermart_internal_app_gophermartapi_models.Order": {
            "type": "object",
            "properties": {
                "accrual": {
                    "type": "number"
                },
                "number": {
                    "type": "string",
                    "example": "0"
                },
                "status": {
                    "type": "string"
                },
                "uploaded_at": {
                    "type": "string"
                }
            }
        },
        "github_com_NStegura_gophermart_internal_app_gophermartapi_models.User": {
            "type": "object",
            "properties": {
                "login": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "github_com_NStegura_gophermart_internal_app_gophermartapi_models.WithdrawIn": {
            "type": "object",
            "properties": {
                "order": {
                    "type": "string"
                },
                "sum": {
                    "type": "number"
                }
            }
        },
        "github_com_NStegura_gophermart_internal_app_gophermartapi_models.WithdrawOut": {
            "type": "object",
            "properties": {
                "order": {
                    "type": "string"
                },
                "processed_at": {
                    "type": "string"
                },
                "sum": {
                    "type": "number"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Gophermart API",
	Description:      "This is a Gophermart server.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
