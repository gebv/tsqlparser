-- Page name: 'Просмотр грузовых автомобилей'
--
-- Hash: 3ad0b4de93c4b3a89292850360225737
-- Seq: #0
-- DB: forddb
-- SQLFormat: select null as ww, lexicon0_.DescriptionId as col_0_0_, lexicon0_.LanguageId as col_1_0_, lexicon0_.MarketId as col_2_0_, lexicon0_.Description as col_3_0_ from Lexicon lexicon0_ where lexicon0_.SourceId=@P0 and (lexicon0_.DescriptionId in (@P1 , @P2 , @P3 , @P4 , @P5 , @P6 , @P7 , @P8 , @P9 , @P10 , @P11 , @P12 , @P13)) and lexicon0_.LanguageId=@P14 and (lexicon0_.MarketId in (@P15 , @P16)) order by lexicon0_.DescriptionId ASC, lexicon0_.LanguageId DESC, lexicon0_.MarketId ASC
-- SQLParams: 4,462170,456949,454600,454569,458481,461081,462196,455610,455509,454911,456943,458710,454187,15,39,1000
-- Description: 
-- 
-- 
USE forddb;
select null as ww, lexicon0_.DescriptionId as col_0_0_, lexicon0_.LanguageId as col_1_0_, lexicon0_.MarketId as col_2_0_, lexicon0_.Description as col_3_0_ from Lexicon lexicon0_ where lexicon0_.SourceId=4 and (lexicon0_.DescriptionId in (462170 , 456949 , 454600 , 454569 , 458481 , 461081 , 462196 , 455610 , 455509 , 454911 , 456943 , 458710 , 454187)) and lexicon0_.LanguageId=15 and (lexicon0_.MarketId in (39 , 1000)) order by lexicon0_.DescriptionId ASC, lexicon0_.LanguageId DESC, lexicon0_.MarketId ASC;


-- 
-- Total rows: 13
-- 
-- Columns: ww	col_0_0_	col_1_0_	col_2_0_	col_3_0_
-- 
-- \N	454187	15	1000	Защелка Крышки Вещевого Отсека
-- \N	454569	15	1000	Гайка УдлинителЯ
-- \N	454600	15	1000	Накладка Порога Двери
-- \N	454911	15	1000	Держатель Цилиндра Замка
-- \N	455509	15	1000	ПетлЯ
-- \N	455610	15	1000	БоковаЯ Отделка Капота
-- \N	456943	15	1000	Винт
-- \N	456949	15	1000	ШестиграннаЯ Гайка
-- \N	458481	15	1000	Замок
-- \N	458710	15	1000	Пластина
-- \N	461081	15	1000	Крышка Экрана Эл.Проводки
-- \N	462170	15	1000	Кронштейн Панели Отделки
-- \N	462196	15	1000	Крышка Домкрата
-- 