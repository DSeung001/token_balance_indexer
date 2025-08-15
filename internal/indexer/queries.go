package indexer

const QBlocks = `
query($gt:Int!, $lt:Int!){
  getBlocks(where:{height:{gt:$gt, lt:$lt}}){
    hash height time num_txs total_txs
  }
}`

const QTxs = `
query($gt:Int!, $lt:Int!, $imax:Int!){
  getTransactions(where:{
    block_height:{gt:$gt, lt:$lt},
    index:{lt:$imax}
  }){
    index hash success block_height
    gas_fee { amount denom }
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
    hash height time num_txs total_txs
  }
}`
