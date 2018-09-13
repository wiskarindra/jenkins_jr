class CreateInfluencers < ActiveRecord::Migration[5.1]
  def up
    create_table :influencers do |t|
      t.string :name, index: true

      t.timestamps null: false
    end
  end

  def down
    drop_table :influencers
  end
end
