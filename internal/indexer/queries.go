package indexer

const QBlocks = `
query($gt:Int!, $lt:Int!){
  getBlocks(where:{height:{gt:$gt, lt:$lt}}){
    hash height last_block_hash time num_txs total_txs
  }
}`

const QTxs = `
query($gt:Int!, $lt:Int!, $imax:Int!){
  getTransactions(where:{
    block_height:{gt:$gt, lt:$lt},
    index:{lt:$imax}
  }){
    index hash success block_height gas_wanted gas_used memo content_raw
    gas_fee { amount denom }
    messages {
      route
      value {
        __typename
      }
    }
    response {
      events {
        ... on GnoEvent {
          type func pkg_path
          attrs { key value }
        }
      }
    }
  }
}`

const QLatestBlock = `
query {
  getBlocks(where:{}, limit: 1, orderBy: {height: DESC}) {
    hash height last_block_hash time num_txs total_txs
  }
}`
