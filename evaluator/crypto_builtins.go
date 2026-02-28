package evaluator

import (
	"base/object"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/google/uuid"
)

func RegisterCryptoBuiltins() {
	builtins["crypto.uuid"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			return &object.String{Value: uuid.New().String()}
		},
	}

	builtins["crypto.hash"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			msg := args[0].Inspect()
			hash := sha256.Sum256([]byte(msg))
			return &object.String{Value: fmt.Sprintf("%x", hash)}
		},
	}

	builtins["encode.base64"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `encode.base64` must be STRING")
			}
			return &object.String{Value: base64.StdEncoding.EncodeToString([]byte(str.Value))}
		},
	}

	builtins["decode.base64"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `decode.base64` must be STRING")
			}
			decoded, err := base64.StdEncoding.DecodeString(str.Value)
			if err != nil {
				return newError("base64 decode error: %s", err.Error())
			}
			return &object.String{Value: string(decoded)}
		},
	}

	builtins["crypto.encrypt_file"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 3 {
				return newError("wrong number of arguments. got=%d, want=3 (algorithm, filepath, key)", len(args))
			}
			filePath, ok1 := args[1].(*object.String)
			keyStr, ok2 := args[2].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `crypto.encrypt_file` must be (STRING, STRING, STRING)")
			}

			plaintext, err := ioutil.ReadFile(filePath.Value)
			if err != nil {
				return newError("could not read file: %s", err.Error())
			}

			keyHash := sha256.Sum256([]byte(keyStr.Value))
			block, err := aes.NewCipher(keyHash[:])
			if err != nil {
				return newError("cipher error: %s", err.Error())
			}

			aesGCM, err := cipher.NewGCM(block)
			if err != nil {
				return newError("gcm error: %s", err.Error())
			}

			nonce := make([]byte, aesGCM.NonceSize())
			if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
				return newError("nonce error: %s", err.Error())
			}

			ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
			encoded := base64.StdEncoding.EncodeToString(ciphertext)
			return &object.String{Value: encoded}
		},
	}

	builtins["crypto.decrypt_file"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 3 {
				return newError("wrong number of arguments. got=%d, want=3 (algorithm, encrypted_data, key)", len(args))
			}
			dataStr, ok1 := args[1].(*object.String)
			keyStr, ok2 := args[2].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `crypto.decrypt_file` must be (STRING, STRING, STRING)")
			}

			ciphertext, err := base64.StdEncoding.DecodeString(dataStr.Value)
			if err != nil {
				return newError("base64 decode error: %s", err.Error())
			}

			keyHash := sha256.Sum256([]byte(keyStr.Value))
			block, err := aes.NewCipher(keyHash[:])
			if err != nil {
				return newError("cipher error: %s", err.Error())
			}

			aesGCM, err := cipher.NewGCM(block)
			if err != nil {
				return newError("gcm error: %s", err.Error())
			}

			nonceSize := aesGCM.NonceSize()
			if len(ciphertext) < nonceSize {
				return newError("ciphertext too short")
			}

			nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
			plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				return newError("decryption error: %s", err.Error())
			}

			return &object.String{Value: string(plaintext)}
		},
	}
}
