class AddInfluencerIdToPosts < ActiveRecord::Migration[5.1]
  def change
    add_column :posts, :influencer_id, 'INT(11) DEFAULT 0'

    add_index :posts, [:influencer_id, :deleted, :published]
  end
end
