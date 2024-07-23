## tg_ai_service

### 1. 目的

通过tg为api入口，将自己的信息流的处理进行ai加持的，提高自己整体信息流的效率；

### 2. 主要的功能

#### 2.1 命令支持

* `/st type`: 存储命令，将对应的类型进行存储；type的类型包含有: file,photo,audio, video, url(转化成markdown进行保存)
* `/ts url|text`: 对url指向的文章或者对应的文字进行翻译的；
* `/a2t`: 将对应的audio转化成文字输出，以md的方式进行输出
* `/sm type url|pdf|text`: 对对应的url、pdf、text进行总结的，后期可能会支持更加多的类型
* `/tor url <collection> <tags>`: 将当前的url保存到raindrop中去;
* `/query content`: 做一个简单的查询问题，
* `/geneimage <model> <desc>`: 根据描述来生成对应的图片
* `/stock <op> <stock code>`: 操作某一支股票的

#### 2.2 包含的模块

* `raindrop-service`: 用来对raindrop进行crud的操作，尤其是添加url，和对raindrop的url进行ai总结和tags生成
* `ai-service`:这个模块主要是封装不同ai模型的，比如openai、gemini，还有其他的模型的，用来进行查询服务
* `storage-service`: 这部分提供的存储有两种，一种是mongodb,一种是对象存储的支持，尤其是对象存储；准备通过opendal
* `url2md-service`:将url变成md的服务，这个其实个一个httpclient服务，通过调用别人的服务将url变成md
* `pdf-service`: 将pdf变成文字的服务，也是一个httpclient服务，调用别人的服务而已
* `podcast-service`: 目前没想法
* `stock-service`: 目前没想法