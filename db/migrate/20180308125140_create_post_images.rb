class CreatePostImages < ActiveRecord::Migration[5.1]
  def up
    create_table :post_images do |t|
      t.integer :post_id, index: true
      t.string  :url
      t.integer :height
      t.integer :width

      t.timestamps null: false
    end
  end

  def down
    drop_table :post_images
  end
end
