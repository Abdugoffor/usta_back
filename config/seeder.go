package config

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

func SeedVacanciesAndResumes() {
	ctx := context.Background()

	firstNames := []string{
		"Ali", "Vali", "Hasan", "Husan", "Jasur", "Sardor", "Dilshod", "Bekzod", "Aziz", "Sherzod",
		"Abdulloh", "Muhammad", "Oybek", "Rustam", "Shoxrux", "Nodir", "Akmal", "Sanjar", "Umid", "Temur",
	}

	lastNames := []string{
		"Karimov", "Aliyev", "Raximov", "Qodirov", "Abdullayev", "Tursunov", "Ergashev", "Sattorov", "Bozorov", "Yoqubov",
	}

	resumeTitles := []string{
		"Elektrik ustasi", "Santexnik", "Payvandchi", "Kafelchi", "Mebel ustasi",
		"Tom yopuvchi", "Bo'yoqchi", "G'isht teruvchi", "Betonchi", "Konditsioner ustasi",
	}

	vacancyTitles := []string{
		"Elektrik kerak", "Santexnik kerak", "Payvandchi kerak", "Kafelchi kerak",
		"Mebel ustasi kerak", "Tom yopuvchi kerak", "Bo'yoqchi kerak",
	}

	skillsList := []string{
		"Elektr montaj", "Suv tizimi", "Payvandlash", "Kafel terish",
		"Mebel yig'ish", "Tom yopish", "Bo'yash", "Beton ishlari",
	}

	addresses := []string{
		"Toshkent shahar Chilonzor tumani",
		"Toshkent shahar Yunusobod tumani",
		"Samarqand shahar",
		"Andijon shahar",
		"Namangan shahar",
		"Farg'ona shahar",
	}

	rand.Seed(time.Now().UnixNano())

	// RESUMES
	for i := 1; i <= 1000000; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		fullName := firstName + " " + lastName

		title := resumeTitles[rand.Intn(len(resumeTitles))]
		address := addresses[rand.Intn(len(addresses))]
		skill1 := skillsList[rand.Intn(len(skillsList))]
		skill2 := skillsList[rand.Intn(len(skillsList))]

		slug := fmt.Sprintf("resume-%d-%d", i, time.Now().UnixNano())

		var resumeID int64

		err := DB.QueryRow(ctx, `
			INSERT INTO resumes (
				slug,
				user_id,
				region_id,
				district_id,
				mahalla_id,
				adress,
				name,
				photo,
				title,
				text,
				contact,
				price,
				experience_year,
				skills,
				views_count,
				is_active,
				created_at,
				updated_at
			)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, $16, NOW(), NOW()
			)
			RETURNING id
		`,
			slug,
			rand.Intn(100)+1,
			rand.Intn(14)+1,
			rand.Intn(200)+1,
			rand.Intn(500)+1,
			address,
			fullName,
			fmt.Sprintf("/uploads/resumes/%d.jpg", i),
			title,
			fmt.Sprintf("%s bo'yicha %d yillik tajribaga ega ustaman", title, rand.Intn(15)+1),
			fmt.Sprintf("+99890%07d", rand.Intn(9999999)),
			(rand.Intn(200)+50)*10000,
			rand.Intn(15)+1,
			strings.Join([]string{skill1, skill2}, ", "),
			rand.Intn(10000),
			rand.Intn(100) > 10,
		).Scan(&resumeID)

		if err != nil {
			log.Println("❌ Resume insert error:", err)
			continue
		}

		categoryID := rand.Intn(3) + 1

		_, err = DB.Exec(ctx, `
			INSERT INTO category_resume (
				categorya_id,
				resume_id,
				created_at
			)
			VALUES ($1, $2, NOW())
			ON CONFLICT (categorya_id, resume_id) DO NOTHING
		`, categoryID, resumeID)

		if err != nil {
			log.Println("❌ category_resume insert error:", err)
		}
	}

	// VACANCIES
	for i := 1; i <= 5000; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		fullName := firstName + " " + lastName

		title := vacancyTitles[rand.Intn(len(vacancyTitles))]
		address := addresses[rand.Intn(len(addresses))]
		slug := fmt.Sprintf("vacancy-%d-%d", i, time.Now().UnixNano())

		_, err := DB.Exec(ctx, `
			INSERT INTO vacancies (
				slug,
				user_id,
				region_id,
				district_id,
				mahalla_id,
				adress,
				name,
				title,
				text,
				contact,
				price,
				views_count,
				is_active,
				created_at,
				updated_at
			)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, NOW(), NOW()
			)
		`,
			slug,
			rand.Intn(100)+1,
			rand.Intn(14)+1,
			rand.Intn(200)+1,
			rand.Intn(500)+1,
			address,
			fullName,
			title,
			fmt.Sprintf("%s uchun tajribali odam kerak", title),
			fmt.Sprintf("+99891%07d", rand.Intn(9999999)),
			(rand.Intn(300)+100)*10000,
			rand.Intn(20000),
			rand.Intn(100) > 10,
		)

		if err != nil {
			log.Println("❌ Vacancy insert error:", err)
		}
	}

	log.Println("✅ 5000 ta resume va 5000 ta vacancy qo‘shildi")
}
