# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# Note that this schema.rb definition is the authoritative source for your
# database schema. If you need to create the application database on another
# system, you should be using db:schema:load, not running all the migrations
# from scratch. The latter is a flawed and unsustainable approach (the more migrations
# you'll amass, the slower it'll run and the greater likelihood for issues).
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema.define(version: 20180725130653) do

  create_table "action_log_histories", force: :cascade, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8" do |t|
    t.integer "record_id"
    t.string "record_type"
    t.text "changes"
    t.bigint "actor_id"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["record_id", "record_type"], name: "index_action_log_histories_on_record"
  end

  create_table "categories", force: :cascade, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8" do |t|
    t.integer "bukalapak_category_id", default: 0
    t.string "bukalapak_category_name"
    t.integer "count", default: 0
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["bukalapak_category_id"], name: "index_categories_on_bukalapak_category_id", unique: true
  end

  create_table "influencers", force: :cascade, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8" do |t|
    t.string "name"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["name"], name: "index_influencers_on_name"
  end

  create_table "post_filters", force: :cascade, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8" do |t|
    t.integer "post_id"
    t.integer "bukalapak_category_id", default: 0
    t.index ["post_id", "bukalapak_category_id"], name: "index_post_filters_on_post_id_and_bukalapak_category_id", unique: true
  end

  create_table "post_images", force: :cascade, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8" do |t|
    t.integer "post_id"
    t.string "url"
    t.integer "height"
    t.integer "width"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "position", default: 0
    t.index ["post_id"], name: "index_post_images_on_post_id"
  end

  create_table "post_likes", force: :cascade, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8" do |t|
    t.integer "post_id"
    t.integer "bukalapak_user_id"
    t.boolean "liked", default: true
    t.index ["post_id", "bukalapak_user_id", "liked"], name: "index_post_likes_on_post_and_user_and_liked"
    t.index ["post_id", "bukalapak_user_id"], name: "index_post_likes_on_post_and_user", unique: true
  end

  create_table "post_tags", force: :cascade, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8" do |t|
    t.integer "post_id"
    t.string "name"
    t.string "url"
    t.float "coord_x", limit: 24
    t.float "coord_y", limit: 24
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.bigint "reference_id"
    t.string "reference_type"
    t.integer "bukalapak_category_id", default: 0
    t.integer "post_image_id"
    t.index ["post_id"], name: "index_post_tags_on_post_id"
  end

  create_table "posts", force: :cascade, options: "ENGINE=InnoDB DEFAULT CHARSET=utf8" do |t|
    t.string "title"
    t.text "description"
    t.string "influencer_name"
    t.boolean "published", default: false
    t.datetime "first_published_at"
    t.datetime "last_published_at"
    t.integer "like_count", default: 0
    t.boolean "deleted", default: false
    t.integer "score", default: 0
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "influencer_id", default: 0
    t.index ["deleted", "published"], name: "index_posts_on_deleted_and_published"
    t.index ["influencer_id", "deleted", "published"], name: "index_posts_on_influencer_id_and_deleted_and_published"
  end

end
