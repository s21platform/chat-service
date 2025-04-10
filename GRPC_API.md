# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [api/chat.proto](#api_chat-proto)
    - [Chat](#-Chat)
    - [ChatEmpty](#-ChatEmpty)
    - [CreatePrivateChatIn](#-CreatePrivateChatIn)
    - [CreatePrivateChatOut](#-CreatePrivateChatOut)
    - [DeletePrivateMessageIn](#-DeletePrivateMessageIn)
    - [DeletePrivateMessageOut](#-DeletePrivateMessageOut)
    - [EditPrivateMessageIn](#-EditPrivateMessageIn)
    - [EditPrivateMessageOut](#-EditPrivateMessageOut)
    - [GetChatsOut](#-GetChatsOut)
    - [GetPrivateRecentMessagesIn](#-GetPrivateRecentMessagesIn)
    - [GetPrivateRecentMessagesOut](#-GetPrivateRecentMessagesOut)
    - [Message](#-Message)
  
    - [ChatService](#-ChatService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="api_chat-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/chat.proto



<a name="-Chat"></a>

### Chat



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| last_message | [string](#string) |  | Контент последнего сообщения |
| chat_name | [string](#string) |  | Название чата |
| avatar_url | [string](#string) |  | Аватарка чата |
| last_message_timestamp | [string](#string) |  | Время отправки последнего сообщения |
| chat_uuid | [string](#string) |  | UUID чата |






<a name="-ChatEmpty"></a>

### ChatEmpty







<a name="-CreatePrivateChatIn"></a>

### CreatePrivateChatIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| companion_uuid | [string](#string) |  | uuid второго пользователя, с которым будет идти переписка |






<a name="-CreatePrivateChatOut"></a>

### CreatePrivateChatOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| new_chat_uuid | [string](#string) |  | uuid созданного чата |






<a name="-DeletePrivateMessageIn"></a>

### DeletePrivateMessageIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| chat_uuid | [string](#string) |  | uuid чата |
| message_uuid | [string](#string) |  | uuid сообщения |
| mode | [string](#string) |  | тип удаления: у всех или у себя |






<a name="-DeletePrivateMessageOut"></a>

### DeletePrivateMessageOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deletion_status | [bool](#bool) |  | статус удаления |






<a name="-EditPrivateMessageIn"></a>

### EditPrivateMessageIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| chat_uuid | [string](#string) |  | uuid чата |
| message_uuid | [string](#string) |  | uuid сообщения |
| new_content | [string](#string) |  | новый текст сообщения |






<a name="-EditPrivateMessageOut"></a>

### EditPrivateMessageOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message_uuid | [string](#string) |  | uuid измененного сообщения |
| new_content | [string](#string) |  | новый текст сообщения |
| updated_at | [string](#string) |  | время обновления сообщения |






<a name="-GetChatsOut"></a>

### GetChatsOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| chats | [Chat](#Chat) | repeated | Список чатов |






<a name="-GetPrivateRecentMessagesIn"></a>

### GetPrivateRecentMessagesIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| chat_uuid | [string](#string) |  | uuid чата, из которого достаем сообщения |






<a name="-GetPrivateRecentMessagesOut"></a>

### GetPrivateRecentMessagesOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| messages | [Message](#Message) | repeated | список сообщений |






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





 

 

 


<a name="-ChatService"></a>

### ChatService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreatePrivateChat | [.CreatePrivateChatIn](#CreatePrivateChatIn) | [.CreatePrivateChatOut](#CreatePrivateChatOut) |  |
| GetChats | [.ChatEmpty](#ChatEmpty) | [.GetChatsOut](#GetChatsOut) |  |
| GetPrivateRecentMessages | [.GetPrivateRecentMessagesIn](#GetPrivateRecentMessagesIn) | [.GetPrivateRecentMessagesOut](#GetPrivateRecentMessagesOut) |  |
| DeletePrivateMessage | [.DeletePrivateMessageIn](#DeletePrivateMessageIn) | [.DeletePrivateMessageOut](#DeletePrivateMessageOut) |  |
| EditPrivateMessage | [.EditPrivateMessageIn](#EditPrivateMessageIn) | [.EditPrivateMessageOut](#EditPrivateMessageOut) |  |

 



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

