# External Sources Description

This document describes the algorithm for generating the external value is used in each pulse generated by Random UChile's Beacon.

## 1. Earthquakes
- **External Source**: from the web of Centro Sismologico Nacional (https://www.sismologia.cl), we obtain the information from the last earthquake with a magnitude greater than 2.5 produced in Chilean territory. The collected fields are UTC Date, Latitude, Longitude, Depth, and Magnitude.
- **Intended update frequency**: every minute.
- **Intended update moment**: at the beginning of each minute (0" second mark).
- **Repeat until new available value**: yes.
- **Local hashing**: yes. 
- **Default URL for access**: https://www.sismologia.cl/sismicidad/catalogo/[YYYY]/[MM]/[YYYY][MM][DD].html (where [YYYY], [MM], and [DD], is the year, month and day of the pulse where the event will be used), then need to check the last earthquake produced BEFORE the date of the pulse.
- **Recommended sampling trials**: immediately perform successive sampling.
- **Fallback options**: None.

## 2. Radio
- **External Source**: obtains the byte-stream (mp3-encoded) produced by the public signal of Radio Universidad de Chile(https://radio.uchile.cl/senal-en-vivo/).
- **Intended update frequency**: every minute. 
- **Intended update moment**: at the beginning of each minute (0" second mark), the collector obtains 99840 bytes of the stream (approx. 5 seconds). The extracted data satisfies that the hash of its first frame (using SHA3-512) starts with 00. This defines a transparent criterion that allows any external party to replicate the collection in parallel to the operation of our collector.
- **Repeat until new available value**: yes.
- **Local hashing**: yes. 
- **Default URL for access**: https://sonic-us.streaming-chile.com/8186/stream (online stream of the radio station). 
- **Recommended sampling trials**: immediately perform successive sampling. 
- **Fallback options**: None.

## 3. Ethereum Blockchain
- **External Source**: Ethereum is a cryptocurrency that uses (as Bitcoin) a Blockchain in order to ensure the transactions performed by the users. This blockchain is generating new blocks approx. every 12 seconds. The collector gets the hash value of the last block published in the blockchain, requesting it from 4 API sources and keeping the first valid response. Those sources are a local node, Infura, Etherscan, and Rivet.
- **Intended update frequency**: every minute. 
- **Intended update moment**: at the beginning of each minute (0" second mark) will gets the hash value of the last block published in the blockchain.
- **Repeat until new available value**: yes.
- **Local hashing**: yes. 
- **Default URL for access**: None (a local node for Ethereum Blockchain is used).
- **Recommended sampling trials**: immediately perform successive sampling. 
- **Fallback options**: Infura https://infura.io/ | Etherscan https://etherscan.io | Rivet https://rivet.cloud/docs/topics/api/index.html.

## 4. Public Transport System of Santiago, Chile
- **External Source**: at every minute obtains the properties of each active bus in Red Metropolitana, the public transport system of Santiago, Chile. This information is directly obtained from the endpoints provided by Directorio de Transporte Público Metropolitano (https://www.dtpm.cl/). For each active bus, the collected fields are: UTC date, license plate, latitude, longitude, speed, route and direction.
- **Intended update frequency**: every minute. 
- **Intended update moment**: at the beginning of each minute (0" second mark), the collector will get the properties of every active bus. In this context, we define as "active" any bus that shared its GPS location in the last 2 minutes.
- **Repeat until new available value**: yes.
- **Local hashing**: yes. 
- **Default URL for access**: https://www.dtpm.cl/index.php/sistema-transporte-publico-santiago/datos-y-servicios (here you can request access to the endpoint).
- **Recommended sampling trials**: immediately perform successive sampling. 
- **Fallback options**: None.


All the hashed values obtained, in parallel, from the 4 sources described before will be concatenated and be used as input of a VDF (Verifiable Delay Function) whose output will be the external value used in the pulse generation of the current minute.
The implemented VDF corresponds to ChiaVDF, based on repeated exponentiations in the field of Binary Quadratic Forms, and derived from the work of Benjamin Wesolowski. The used parameters are the folowing:
- **Security Parameter (lambda)**: 1024
- **Exponentiations (T)**: 2000000
- **Seed**: Randomly chosen at the beginning of each minute (0" second mark).
