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
    saiu_de text,
    criado_em timestamptz not null default now()
);
-- o create acima não acrescenta coluna em tabela que já existe.
alter table leads add column if not exists saiu_de text;

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

-- pedidos de coleta feitos pelo painel; o coletor de plantão (coleta --servir) os pega um por vez.
create table if not exists pedidos (
    id bigserial primary key,
    cidade text not null,
    bairro text not null default '',
    nicho text not null,
    limite int not null check (limite between 1 and 20),
    estado text not null default 'pendente' check (estado in ('pendente', 'rodando', 'pronto', 'erro')),
    novos int,
    erro text,
    criado_em timestamptz not null default now(),
    terminado_em timestamptz
);
alter table leads add column if not exists cidade text not null default 'João Pessoa';
alter table leads add column if not exists pedido_id bigint references pedidos (id);
