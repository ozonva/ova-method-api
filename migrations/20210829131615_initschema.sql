-- +goose Up
-- +goose StatementBegin
create table methods
(
    id            bigserial     primary key,
    user_id       bigint        not null,
    value         varchar(255)  not null,
    created_at    timestamp     not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table methods;
-- +goose StatementEnd
