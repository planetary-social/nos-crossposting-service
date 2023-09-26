package firestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"google.golang.org/api/iterator"
)

const (
	collectionPublicKeys               = "publicKeys"
	collectionPublicKeysFieldPublicKey = "publicKey"

	collectionPublicKeysAPNSTokens                      = "apnsTokens"
	collectionPublicKeysAPNSTokensFieldToken            = "token"
	collectionPublicKeysAPNSTokensFieldUpdatedTimestamp = "updatedTimestamp"
)

type PublicKeyRepository struct {
	client *firestore.Client
	tx     *firestore.Transaction
}

func NewPublicKeyRepository(client *firestore.Client, tx *firestore.Transaction) *PublicKeyRepository {
	return &PublicKeyRepository{client: client, tx: tx}
}

func (r *PublicKeyRepository) Save(registration domain.Registration) error {
	pubKeyDocPath := r.client.Collection(collectionPublicKeys).Doc(registration.PublicKey().Hex())
	pubKeyDocData := map[string]any{
		collectionPublicKeysFieldPublicKey: ensureType[string](registration.PublicKey().Hex()),
	}
	if err := r.tx.Set(pubKeyDocPath, pubKeyDocData, firestore.MergeAll); err != nil {
		return errors.Wrap(err, "error creating the public key doc")
	}

	tokenDocPath := r.client.Collection(collectionPublicKeys).Doc(registration.PublicKey().Hex()).Collection(collectionPublicKeysAPNSTokens).Doc(registration.APNSToken().Hex())
	tokenDocData := map[string]any{
		collectionPublicKeysAPNSTokensFieldToken:            ensureType[string](registration.APNSToken().Hex()),
		collectionPublicKeysAPNSTokensFieldUpdatedTimestamp: ensureType[time.Time](time.Now()),
	}
	if err := r.tx.Set(tokenDocPath, tokenDocData, firestore.MergeAll); err != nil {
		return errors.Wrap(err, "error creating the public key doc")
	}

	return nil
}

func (r *PublicKeyRepository) GetAPNSTokens(ctx context.Context, publicKey domain.PublicKey, savedAfter time.Time) ([]domain.APNSToken, error) {
	docs := r.tx.Documents(
		r.client.
			Collection(collectionPublicKeys).
			Doc(publicKey.Hex()).
			Collection(collectionPublicKeysAPNSTokens).
			Where(collectionPublicKeysAPNSTokensFieldUpdatedTimestamp, ">", savedAfter),
	)

	var result []domain.APNSToken

	for {
		doc, err := docs.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, errors.Wrap(err, "error getting a document")
		}

		data := make(map[string]any)

		if err := doc.DataTo(&data); err != nil {
			return nil, errors.Wrap(err, "error reading document data")
		}

		apnsToken, err := domain.NewAPNSTokenFromHex(data["token"].(string))
		if err != nil {
			return nil, errors.Wrap(err, "error creating a token from hex")
		}

		result = append(result, apnsToken)
	}

	return result, nil
}
