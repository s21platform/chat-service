# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [api/chat.proto](#api_chat-proto)
    - [ChatUser](#-ChatUser)
    - [CreateStreamIn](#-CreateStreamIn)
    - [CreateStreamOut](#-CreateStreamOut)
    - [GetBatchSubscribeTokensIn](#-GetBatchSubscribeTokensIn)
    - [GetBatchSubscribeTokensOut](#-GetBatchSubscribeTokensOut)
    - [GetConnectAccessTokenOut](#-GetConnectAccessTokenOut)
    - [GetPrivateStreamsOut](#-GetPrivateStreamsOut)
    - [GetStreamRecentMessagesIn](#-GetStreamRecentMessagesIn)
    - [GetStreamRecentMessagesOut](#-GetStreamRecentMessagesOut)
    - [GetStreamSubscribeTokenIn](#-GetStreamSubscribeTokenIn)
    - [GetStreamSubscribeTokenOut](#-GetStreamSubscribeTokenOut)
    - [GetUserActiveStreamsOut](#-GetUserActiveStreamsOut)
    - [Message](#-Message)
    - [PrivateStream](#-PrivateStream)
    - [SendMessageIn](#-SendMessageIn)
    - [SendMessageOut](#-SendMessageOut)
    - [StreamSubscription](#-StreamSubscription)
  
    - [ChatService](#-ChatService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="api_chat-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/chat.proto



<a name="-ChatUser"></a>

### ChatUser



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | id пользователя |
| metadata | [string](#string) |  | метаданные пользователя |






<a name="-CreateStreamIn"></a>

### CreateStreamIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [ChatUser](#ChatUser) | repeated | пользователи стрима |
| type | [string](#string) |  | тип стрима |
| chat_metadata | [string](#string) |  | метаданные чата |
| creator_metadata | [string](#string) |  | метаданные создателя |






<a name="-CreateStreamOut"></a>

### CreateStreamOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | id созданного стрима |






<a name="-GetBatchSubscribeTokensIn"></a>

### GetBatchSubscribeTokensIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stream_ids | [string](#string) | repeated | Список ID стримов для которых нужны токены |






<a name="-GetBatchSubscribeTokensOut"></a>

### GetBatchSubscribeTokensOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| subscriptions | [StreamSubscription](#StreamSubscription) | repeated | Токены для подписки на стримы |






<a name="-GetConnectAccessTokenOut"></a>

### GetConnectAccessTokenOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [string](#string) |  | JWT токен для подключения к Centrifugo |
| expires_at | [int64](#int64) |  | время истечения токена (unix timestamp) |






<a name="-GetPrivateStreamsOut"></a>

### GetPrivateStreamsOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| streams | [PrivateStream](#PrivateStream) | repeated | Список стримов |






<a name="-GetStreamRecentMessagesIn"></a>

### GetStreamRecentMessagesIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stream_id | [string](#string) |  | ID стрима, из которого достаем сообщения |
| offset | [string](#string) |  | временная метка для offset (RFC3339) |
| limit | [int32](#int32) |  | количество сообщений для возврата |






<a name="-GetStreamRecentMessagesOut"></a>

### GetStreamRecentMessagesOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| messages | [Message](#Message) | repeated | список сообщений |






<a name="-GetStreamSubscribeTokenIn"></a>

### GetStreamSubscribeTokenIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stream_id | [string](#string) |  | ID стрима для получения subscribe токена |






<a name="-GetStreamSubscribeTokenOut"></a>

### GetStreamSubscribeTokenOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [string](#string) |  | JWT токен для подписки на канал стрима |
| expires_at | [int64](#int64) |  | время истечения токена (unix timestamp) |
| channel | [string](#string) |  | канал для подписки в Centrifugo |






<a name="-GetUserActiveStreamsOut"></a>

### GetUserActiveStreamsOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stream_ids | [string](#string) | repeated | Список ID стримов где пользователь является участником |






<a name="-Message"></a>

### Message



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uuid | [string](#string) |  | uuid пользователя |
| content | [string](#string) |  | само сообщение |
| sent_at | [string](#string) |  | время отправки |
| updated_at | [string](#string) |  | время обновления |
| root_uuid | [string](#string) |  | uuid корневого сообщения |
| parent_uuid | [string](#string) |  | uuid сообщения, на которое идет прямой ответ |






<a name="-PrivateStream"></a>

### PrivateStream



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stream_id | [string](#string) |  | ID стрима |
| last_message_content | [string](#string) |  | Контент последнего сообщения |
| stream_name | [string](#string) |  | Название стрима |
| avatar_url | [string](#string) |  | Аватарка стрима |
| last_message_timestamp | [string](#string) |  | Время отправки последнего сообщения |






<a name="-SendMessageIn"></a>

### SendMessageIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stream_id | [string](#string) |  | ID стрима |
| content | [string](#string) |  | текст сообщения |
| parent_id | [string](#string) | optional | ID сообщения для reply (опционально) |
| root_id | [string](#string) | optional | ID корневого сообщения треда (опционально) |
| message_type | [string](#string) |  | тип сообщения (text, image, file) |






<a name="-SendMessageOut"></a>

### SendMessageOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message_id | [string](#string) |  | ID созданного сообщения |
| sent_at | [string](#string) |  | время отправки |






<a name="-StreamSubscription"></a>

### StreamSubscription



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stream_id | [string](#string) |  | ID стрима |
| token | [string](#string) |  | JWT токен для подписки |
| expires_at | [int64](#int64) |  | время истечения токена |
| channel | [string](#string) |  | канал для подписки в Centrifugo |





 

 

 


<a name="-ChatService"></a>

### ChatService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateStream | [.CreateStreamIn](#CreateStreamIn) | [.CreateStreamOut](#CreateStreamOut) |  |
| SendMessage | [.SendMessageIn](#SendMessageIn) | [.SendMessageOut](#SendMessageOut) |  |
| GetPrivateStreams | [.google.protobuf.Empty](#google-protobuf-Empty) | [.GetPrivateStreamsOut](#GetPrivateStreamsOut) |  |
| GetStreamRecentMessages | [.GetStreamRecentMessagesIn](#GetStreamRecentMessagesIn) | [.GetStreamRecentMessagesOut](#GetStreamRecentMessagesOut) |  |
| GetConnectAccessToken | [.google.protobuf.Empty](#google-protobuf-Empty) | [.GetConnectAccessTokenOut](#GetConnectAccessTokenOut) |  |
| GetStreamSubscribeToken | [.GetStreamSubscribeTokenIn](#GetStreamSubscribeTokenIn) | [.GetStreamSubscribeTokenOut](#GetStreamSubscribeTokenOut) |  |
| GetUserActiveStreams | [.google.protobuf.Empty](#google-protobuf-Empty) | [.GetUserActiveStreamsOut](#GetUserActiveStreamsOut) |  |
| GetBatchSubscribeTokens | [.GetBatchSubscribeTokensIn](#GetBatchSubscribeTokensIn) | [.GetBatchSubscribeTokensOut](#GetBatchSubscribeTokensOut) |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

