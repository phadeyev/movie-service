package schema

import (
	"github.com/jmoiron/sqlx"
)

// seeds is a string constant containing all of the queries needed to get the
// db seeded to a useful state for development.
//
// Using a constant in a .go file is an easy way to ensure the queries are part
// of the compiled executable and avoids pathing issues with the working
// directory. It has the downside that it lacks syntax highlighting and may be
// harder to read for some cases compared to using .sql files. You may also
// consider a combined approach using a tool like packr or go-bindata.
//
// Note that database servers besides PostgreSQL may not support running
// multiple queries as part of the same execution so this single large constant
// may need to be broken up.

const seeds = `
INSERT INTO movies (movie_id, name, description, director, year, ageRating, poster, youtubeVideoId, isPaid) VALUES
	('a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 'Бойцовский клуб', 'Американский кинофильм 1999 года режиссёра Дэвида Финчера по мотивам одноимённого романа Чака Паланика, вышедшего тремя годами ранее. Главные роли исполняют Эдвард Нортон, Брэд Питт и Хелена Бонэм Картер. Нортон исполняет роль безымянного рассказчика — обезличенного обывателя, который недоволен своей жизнью в постиндустриальном потребительском обществе «белых воротничков». Он создаёт подпольную организацию «Бойцовский клуб» вместе с Тайлером Дёрденом — продавцом мыла, роль которого исполнил Брэд Питт.', 'Дэвид Финчер', 1999, 18, '/static/posters/fightclub.jpg', 'qtRKdVHc-cE', true),
	('72f8b983-3eb4-48db-9ed0-e45cc6bd716b', 'Крестный отец', 'Эпическая гангстерская драма режиссёра Фрэнсиса Форда Копполы. Экранизация одноимённого романа Марио Пьюзо, изданного в 1969 году. Слоган: «Предложение, от которого невозможно отказаться». Главные роли Вито и Майкла Корлеоне исполняют Марлон Брандо и Аль Пачино соответственно. Во второстепенных ролях снялись Джеймс Каан и Роберт Дюваль.', 'Фрэнсис Форд Коппола', '1972', '18', '/static/posters/father.jpg', 'ar1SHxgeZUc', false),
	('efd3c33c-e5e2-11ea-adc1-0242ac120002', 'Криминальное чтиво', 'Кинофильм режиссёра Квентина Тарантино[5]. Сюжет фильма нелинеен, как и почти во всех остальных работах Тарантино. Этот приём стал чрезвычайно популярен, породив множество подражаний во второй половине 1990-х[6]. В фильме рассказывается несколько историй, в которых показаны ограбления, философские дискуссии двух гангстеров, спасение девушки от передозировки героина и боксёр, которого задели за живое. Название является отсылкой к популярным в середине XX века в США pulp-журналам. Именно в стиле таких журналов были оформлены афиши, а позднее саундтрек, видеокассеты и DVD с фильмом.', 'Квентин Тарантино', '1994', '16', '/static/posters/pulpfiction.jpg', 's7EdQ4FqbhY', true)
	ON CONFLICT DO NOTHING;
`

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
