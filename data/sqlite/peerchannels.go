package sqlite

import (
	"context"
	"time"

	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

const sqlPeerChannelAccountSelect = `
	SELECT account_id, user_name, password FROM users_peerchannels
	WHERE user_id = :user_id
`

const sqlPeerChannelInsert = `
	INSERT INTO peerchannels (peerchannels_account_id, channel_id, channel_host, channel_path, channel_type, created_at)
	VALUES (:peerchannels_account_id, :channel_id, :channel_host, :channel_path, :channel_type, :created_at)
`

const sqlPeerChannelsAPITokInsert = `
	INSERT INTO peerchannels_toks (peerchannels_channel_id, tok, role, can_read, can_write)
	VALUES (:peerchannels_channel_id, :tok, :role, :can_read, :can_write)
`

const sqlPeerChannelsCloseUpdate = `
	UPDATE peerchannels
	SET closed=1
	WHERE channel_id = :channel_id
`

const sqlPeerChannelsOpenSelect = `
	SELECT pc.channel_host, pc.channel_path, pc.channel_id, pc.channel_type, pc.created_at, pt.tok
	FROM peerchannels pc
	JOIN peerchannels_toks pt ON pc.channel_id = pt.peerchannels_channel_id
	WHERE pc.closed = 0 AND pc.channel_type = :channel_type
`

func (s *sqliteStore) PeerChannelAccount(ctx context.Context, args *payd.PeerChannelIDArgs) (*payd.PeerChannelAccount, error) {
	var row struct {
		ID       int64  `db:"account_id"`
		UserName string `db:"user_name"`
		Password string `db:"password"`
	}
	if err := s.db.GetContext(ctx, &row, sqlPeerChannelAccountSelect, args.UserID); err != nil {
		return nil, errors.Wrapf(err, "failed to get peer channel for user id %d", args.UserID)
	}
	return &payd.PeerChannelAccount{
		ID:       row.ID,
		Username: row.UserName,
		Password: row.Password,
	}, nil
}

func (s *sqliteStore) PeerChannelCreate(ctx context.Context, args *payd.PeerChannelCreateArgs) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to create tx for channel %s", args.ChannelID)
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if args.CreatedAt.IsZero() {
		args.CreatedAt = time.Now()
	}
	if err := handleNamedExec(tx, sqlPeerChannelInsert, args); err != nil {
		return errors.Wrapf(err, "failed to insert channel %s", args.ChannelID)
	}
	return errors.Wrapf(commit(ctx, tx), "failed to commit creating channel %s", args.ChannelID)
}

func (s *sqliteStore) PeerChannelAPITokenCreate(ctx context.Context, args *payd.PeerChannelAPITokenStoreArgs) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to create tx for api token %s for channel %s", args.Token, args.PeerChannelsChannelID)
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()

	if err := handleNamedExec(tx, sqlPeerChannelsAPITokInsert, args); err != nil {
		return errors.Wrapf(err, "failed to insert api token %s for channel %s", args.Token, args.PeerChannelsChannelID)
	}
	return errors.Wrapf(commit(ctx, tx), "failed to commit creating token %s for channel %s", args.Token, args.PeerChannelsChannelID)
}

func (s *sqliteStore) PeerChannelAPITokensCreate(ctx context.Context, entries ...*payd.PeerChannelAPITokenStoreArgs) error {
	for _, entry := range entries {
		if err := s.PeerChannelAPITokenCreate(ctx, entry); err != nil {
			return err
		}
	}

	return nil
}

func (s *sqliteStore) PeerChannelsOpened(ctx context.Context, channelType payd.PeerChannelHandlerType) ([]payd.PeerChannel, error) {
	var resp []payd.PeerChannel
	if err := s.db.SelectContext(ctx, &resp, sqlPeerChannelsOpenSelect, channelType); err != nil {
		return nil, errors.Wrap(err, "failed to query for opened peer channels")
	}
	return resp, nil
}

func (s *sqliteStore) PeerChannelCloseChannel(ctx context.Context, channelID string) error {
	if _, err := s.db.NamedExecContext(ctx, sqlPeerChannelsCloseUpdate, struct {
		ChannelID string `db:"channel_id"`
	}{
		ChannelID: channelID,
	}); err != nil {
		return errors.Wrapf(err, "failed to close channel %s", channelID)
	}

	return nil
}
