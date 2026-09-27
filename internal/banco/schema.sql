create table if not exists leads (
    id bigserial primary key,
    place_id text not null unique,
    slug text not null unique,
    nicho text not null,
    bairro text not null,
    nome text not null,
    telefone text not null,
    site text not null default '',
    endereco text not null default '',
    nota numeric not null,
    avaliacoes int,
    dores text[] not null default '{}',
    detalhes jsonb not null default '{}',
    pontuacao int not null,
    variante text not null,
    mensagem text not null,
    etapa text not null default 'na_fila'
        check (etapa in ('na_fila', 'contatado', 'abriu', 'respondeu', 'interessado', 'proposta', 'fechado', 'perdido', 'saiu')),
    contatado_em timestamptz,
    follow_up_em timestamptz,
    respondeu_em timestamptz,
    criado_em timestamptz not null default now()
);

-- só descarte durável (fechado, sem celular, fora do porte, sem dor): descarte por erro não entra aqui e volta a ser tentado.
create table if not exists vistos (
    place_id text primary key,
    motivo text not null,
    visto_em timestamptz not null default now()
);

create table if not exists mensagens (
    id bigserial primary key,
    lead_id bigint not null references leads (id),
    autor text not null check (autor in ('eu', 'lead')),
    texto text not null,
    criado_em timestamptz not null default now()
);

create table if not exists visitas (
    id bigserial primary key,
    slug text not null,
    criado_em timestamptz not null default now()
);

create table if not exists custos (
    dia date not null,
    api text not null,
    unidades bigint not null default 0,
    usd numeric not null default 0,
    primary key (dia, api)
);
